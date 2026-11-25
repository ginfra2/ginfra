/*
Package host
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage hosted deployments.
*/
package host

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	dcontainer "github.com/docker/docker/api/types/container"
	dnetwork "github.com/docker/docker/api/types/network"
	dockerc "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// #####################################################################################################################
// # CONTAINER UTILS

// createContainer creates a Docker container for a hosted service.
//
// Parameters:
//   - hosted: A pointer to a HostFile struct that contains deployment metadata.
//   - cli: A pointer to a Docker client.
//   - image: The Docker image for the container.
//   - port: The primary port for the service.
//   - sspc: Is the service specifics that come from the configuration server or passed to the command line.  This gives
//     additional runtime information to the service that is needed before the configuration provider service is available.
//   - cspc: Is the container specifics that come from the specifics.json file in the container image.  This gives hints
//     to Ginfra about how the container should be run.
//   - cmd: Command line to pass to executable.
//   - dport : Debugger port.  If 0, the port will be ignored.
//   - gpus  : GPU spec.  Right now, if it is not "", it will assign the first GPU to the container.
//
// Returns:
// - string: The ID of the created container.
func createContainer(hosted *manage.HostFile, cli *dockerc.Client, image, port string, sspc map[string]interface{},
	cspc *shared.Specifics, cmd string, dport int, gpus string) string {

	// The primary port for the service.
	var nport = nat.Port(fmt.Sprintf("%s/tcp", port))

	// Port bindings including alternates in the sspc.
	portBindings := make(nat.PortMap)
	portExposed := make(nat.PortSet)

	thp := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: port,
	}
	portBindings[nport] = []nat.PortBinding{thp}
	portExposed[nport] = struct{}{}

	if dport > 0 {
		ndport := nat.Port(fmt.Sprintf("%d/tcp", dport))
		dp := nat.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: strconv.Itoa(dport),
		}
		portBindings[ndport] = []nat.PortBinding{dp}
		portExposed[ndport] = struct{}{}
	}

	// Ports declared by the service in the specifics.json file.
	if cspc != nil {
		for pname := range cspc.Exposed {

			// Scrub the name against the server specifics.
			nn, _ := common.SimpleReplacer(pname, sspc, true)

			p := nat.Port(fmt.Sprintf("%s/tcp", nn))
			hp := nat.PortBinding{
				HostIP:   "0.0.0.0",
				HostPort: pname,
			}
			portBindings[p] = []nat.PortBinding{hp}
			portExposed[p] = struct{}{}
		}
	}

	// Host config
	hcfg := dcontainer.HostConfig{
		PortBindings: portBindings,
	}
	if gpus != "" {
		hcfg.Resources = dcontainer.Resources{
			DeviceRequests: []dcontainer.DeviceRequest{
				{
					Driver:       "nvidia",
					Count:        0,
					DeviceIDs:    []string{"0"},
					Capabilities: [][]string{{"gpu"}},
					Options:      nil,
				},
			},
		}
	}

	resp, err := cli.ContainerCreate(context.Background(),
		&dcontainer.Config{
			Image:        image,
			ExposedPorts: portExposed,
			Cmd:          strings.Fields(cmd),
		},
		&hcfg,
		&dnetwork.NetworkingConfig{
			EndpointsConfig: map[string]*dnetwork.EndpointSettings{hosted.NetworkName: {NetworkID: hosted.NetworkId}},
		},
		nil, "")
	if err != nil {
		panic(base.NewGinfraErrorChild("Could not create container.", err))
	}
	return resp.ID
}

func removeContainer(containerId string, cli *dockerc.Client) error {
	var err error
	if containerId != "" {
		err = cli.ContainerRemove(context.Background(), containerId, dcontainer.RemoveOptions{Force: true})
	}
	return err
}

func rollbackContainer(gctx *app.GContext, containerId string, cli *dockerc.Client) {
	gctx.GetLogger().Info("Rolling back container.")
	if err := removeContainer(containerId, cli); err != nil {
		gctx.GetLogger().Errorw("Could not rollback container.", zap.String(base.LM_NETWORK_ID, containerId),
			err)
	}
}

func injectFile(cli *dockerc.Client, containerId, path, data string) {
	err := cli.CopyToContainer(context.Background(), containerId, path, strings.NewReader(data),
		dcontainer.CopyToContainerOptions{})
	if err != nil {
		panic(base.NewGinfraErrorChild("Could not inject file into container.", err))
	}
}

func retrieveFile(cli *dockerc.Client, containerId, path string) (string, error) {
	re, _, err := cli.CopyFromContainer(context.Background(), containerId, path)
	if err != nil {
		return "", base.NewGinfraErrorChildFlags("Could not read file from container.", err, base.ERRFLAG_NOT_FOUND)
	}

	// create a new tar reader
	tr := tar.NewReader(re)

	header, err := tr.Next()
	if err != nil {
		return "", base.NewGinfraErrorChildFlags("File not available.", err, base.ERRFLAG_NOT_FOUND)
	}
	if header.Typeflag != tar.TypeReg {
		return "", base.NewGinfraErrorChild("Did not get an actual file.", err)
	}

	var sb strings.Builder
	if _, err := io.Copy(&sb, tr); err != nil {
		panic(err) // This should never happen and would be a serious problem.
	}

	return sb.String(), nil
}

// #####################################################################################################################
// # MANAGEMENT UTILS

func rollbackNetwork(gctx *app.GContext, networkId string, cli *dockerc.Client) {
	if networkId != "" {
		gctx.GetLogger().Info("Rolling back network create.")
		if err := cli.NetworkRemove(context.Background(), networkId); err != nil {
			gctx.GetLogger().Errorw("Could not rollback network.", zap.String(base.LM_NETWORK_ID, networkId),
				err)
		}
	}
}

// #####################################################################################################################
// # SPECIFICS UTILS

func getSpecificsCacheFileName(id string, name string) string {
	return strings.ReplaceAll(name, "/", "_") + "_" + strings.ReplaceAll(id, "/", "_")
}

func cacheContainerSpecifics(cli *dockerc.Client, imageId, target string, values map[string]interface{}) (*shared.Specifics, error) {

	// We need to create the container, so we can pull the file.
	resp, err := cli.ContainerCreate(context.Background(),
		&dcontainer.Config{Image: imageId}, nil, nil, nil, "")
	defer func(cli *dockerc.Client, containerID string) {
		if resp.ID != "" {
			_ = cli.ContainerRemove(context.Background(), containerID, dcontainer.RemoveOptions{Force: true})
		}
	}(cli, resp.ID)
	if err != nil {
		return nil, base.NewGinfraErrorChildFlags("Could not get specifics from container.", err,
			base.ERRFLAG_RECOVERABLE)
	}

	// Get the file
	p := filepath.Join("/", app.PathSpecificsFile)
	s, err := retrieveFile(cli, resp.ID, p)
	if err != nil {
		if (base.GetErrorFlags(err) & base.ERRFLAG_NOT_FOUND) > 0 {
			return nil, base.NewGinfraErrorChildAFlags("Specifics file not found in container", err,
				base.ERRFLAG_NOT_FOUND&base.ERRFLAG_RECOVERABLE, base.LM_PATH, p)
		} else {
			return nil, base.NewGinfraErrorChildFlags("Could not get specifics from file", err,
				base.ERRFLAG_RECOVERABLE)
		}
	}

	sp, err := shared.SpecificsLoad(s, p, values)
	if err != nil {
		return nil, base.NewGinfraErrorChildFlags("Specifics file invalid.", err, base.ERRFLAG_RECOVERABLE)
	}

	// Cache it.
	if err = os.WriteFile(target, []byte(s), 0644); err != nil {
		err = base.NewGinfraErrorChildFlags("Could not cache file.", err, base.ERRFLAG_RECOVERABLE)
	}

	return sp, err
}

func getContainerSpecifics(cli *dockerc.Client, image string, values map[string]interface{}) (*shared.Specifics, error) {

	var (
		cspc *shared.Specifics
		err  error
		id   string
	)

	// TODO I'm not sure this is necessary.
	var rr bytes.Buffer
	raw := dockerc.ImageInspectWithRawResponse(&rr)

	// Get current image id for the requested image.
	i, err := cli.ImageInspect(context.Background(), image, raw)
	if err != nil {
		return nil, base.NewGinfraErrorChildA("Unable to access image.", err, base.LM_NAME, image)
	}
	if strings.HasPrefix(i.ID, "sha256:") {
		id = base.SnipId(i.ID[7:])
	} else {
		id = base.SnipId(i.ID)
	}

	// Check cache
	cfile := filepath.Join(base.GetGinfraTmpCacheDir(), getSpecificsCacheFileName(id, image))
	if _, err = os.Stat(cfile); err == nil {
		cspc, err = shared.SpecificsLoadFile(cfile, values)
		if err != nil {
			// Corrupt?  Try to cache it again.
			cspc, err = cacheContainerSpecifics(cli, id, cfile, values)
			if err != nil {
				// Bad thing.
				panic(fmt.Sprintf("Specifics in cache are corrupt and could not be recached.  Clear the cache.  %v", err))
			}
		}

	} else if errors.Is(err, os.ErrNotExist) {
		cspc, err = cacheContainerSpecifics(cli, id, cfile, values)

	} else {
		panic(fmt.Sprintf("Filesystem error.  %v", err))
	}

	return cspc, err
}

// #####################################################################################################################
// # GENERAL UTILS

func mapToTarString(m map[string]interface{}) string {
	d, _ := json.Marshal(m)
	s, err := manage.DataToTar(base.ConfigFileName, d)
	if err != nil {
		panic(base.NewGinfraErrorChild("Could not create config file tar.", err))
	}
	return s
}

// getServicePort TODO This is pretty rickety.  Check to make sure it is a good port.  And there is a potential race condition with the port reservation.
func getServicePort(gctx *app.GContext, cp config.ConfigProvider, hosted *manage.HostFile, name string) int {
	var (
		err error
		p   int
	)

	// Is one reserved.
	c := cp.FetchConfig(config.CR_RESERVATIONS, hosted.DeploymentName)
	if c.Err != nil {
		// No deployment not created yet
		c = config.NewConfig(map[string]interface{}{
			config.CR_RESERVATIONS_PORTS: map[string]interface{}{}}, nil)

	}

	cc := c.GetConfigPath(config.CR_RESERVATIONS_PORTS, name)
	if cc.Err != nil {
		// Reserve one.
		// TODO do I need to switch this to an atomic?
		p = hosted.NextServicePort
		hosted.NextServicePort++
		if err := common.WriteJsonFile[manage.HostFile](*hosted, manage.GetHostFilePath(gctx), "hosting metadata"); err != nil {
			panic(fmt.Sprintf("Could not write metadata file.  This should never happen.  %v", err))
		}

		nc := config.NewConfig(map[string]interface{}{name: strconv.Itoa(p)}, nil)
		if err = c.PutConfigPath(nc, config.CR_RESERVATIONS_PORTS); err != nil {
			panic(fmt.Sprintf("Could not reserve the service port.  %v", err))
		}

	} else {
		// It is already reserved.
		p, err = strconv.Atoi(cc.GetValue(name).GetValueStringNoerr())
		if err != nil {
			panic(fmt.Sprintf("Reserved port is an invalid port value: %v", cc.GetValue(name).GetValueStringNoerr()))
		}

	}

	if err = cp.PostConfig(c, config.CR_RESERVATIONS, hosted.DeploymentName); err != nil {
		panic(fmt.Sprintf("Spurious configuration error.  %v", err))
	}

	return p
}
