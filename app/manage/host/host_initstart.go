/*
Package host
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage hosted deployments which are either local docker container host.

TODO: support external container runtimes.  Right now it is hardcoded to look for a local runtime only.
*/
package host

import (
	"context"
	"fmt"
	"os"
	"strconv"

	dtypes "github.com/docker/docker/api/types"
	dcontainer "github.com/docker/docker/api/types/container"
	dnetwork "github.com/docker/docker/api/types/network"
	dockerc "github.com/docker/docker/client"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"go.uber.org/zap"
)

// #####################################################################################################################
// #

// HostInit initializes a hosted deployment.  The directory it is initialized in will be for this deployment only.
//
// Parameters:
// - gctx: The ginfra app context.
// - configDockerImage: The Docker image for the configuration service container.
//
// Returns:
// - err: An error, if any.
//
// Flow:
// - Checks if the metadata file for the deployment already exists.
// - Gets the Docker client.
// - Creates a network for the deployment.
// - Creates a configuration for starting the configuration provider service.
// - Creates a configuration provider service container.
// - Injects the configuration into the container.
// - Starts the configuration service container.
// - Checks if the configuration provider is running.
// - Retrieves container information.
// - Writes the host metadata file.
// - Logs the successful initialization of the deployment.
func HostInit(gctx *app.GContext, configDockerImage string) (err error) {

	var (
		hosted = manage.HostFile{
			DeploymentName:      gctx.Deployment,
			ConfigAccessToken:   config.GetNewAuthToken(false),
			NextServicePort:     config.DEFAULT_SERVICE_PORT,
			ConfigPortalHostUrl: "http://localhost:" + strconv.Itoa(config.RESERVED_CONFIG_SERVICE_PORT),
			DockerHost:          "", // TODO make configurable.
		}
		cjson dtypes.ContainerJSON
		cli   *dockerc.Client
		cspc  shared.Specifics
	)

	hosted.DockerHost, err = os.Hostname()
	if err != nil {
		panic("Could not get hostname.  " + err.Error())
	}

	if _, err = os.Stat(manage.GetHostFilePath(gctx)); os.IsNotExist(err) == false {
		return base.NewGinfraErrorA("Host deployment metadata file already exists.  Is there another deployment?",
			base.LM_FILEPATH, app.PathHostingFilePath)
	}

	dctx := context.Background()

	// Get client
	// TODO allow this to point to a remote docker engine, if hosted.ConfigPortalHostUrl does not equal local hostname or ip.
	cli, err = dockerc.NewClientWithOpts(dockerc.FromEnv, dockerc.WithAPIVersionNegotiation())
	if err != nil {
		panic(fmt.Sprintf("Could not get docker client: %v", err))
	}
	defer func() {
		_ = cli.Close() // Little late to do anything about it.
	}()

	// Make sure we rollback even if there is a panic from something else.
	defer func() {
		if r := recover(); r != nil {
			if perr, ok := r.(*base.GinfraError); ok { // TODO might be dangerous if another object comes out.
				err = perr
			} else {
				err = base.NewGinfraErrorA("Failed to init host.", base.LM_CAUSE, fmt.Sprintf("%v", r))
			}
			if !gctx.NoRollback {
				rollbackContainer(gctx, hosted.ConfigContainerId, cli)
				rollbackNetwork(gctx, hosted.NetworkId, cli)
			}
		}
	}()

	// -- CREATE NETWORK
	hosted.NetworkName = app.HostNetworkNamePrefix + gctx.Deployment
	var netr dnetwork.CreateResponse
	netr, err = cli.NetworkCreate(dctx, hosted.NetworkName,
		dnetwork.CreateOptions{})
	if err != nil {
		panic(base.NewGinfraErrorChildA("Could not create network.  Can you connect to the docker host or does the network already exist?  Did you forget to run 'ginfra manage host destroy' on previous host deployment?",
			err, base.LM_NETWORK_NAME, hosted.NetworkName))
	}
	hosted.NetworkId = base.SnipId(netr.ID)

	// -- CREATE CONFIG
	var cfg = make(map[string]interface{})
	cfg[base.ConfigAuth] = base.GenerateCodeword(base.ConfigAuthLength)
	cfg[base.ConfigUri] = "config://"
	cfg[base.ConfigPortalUri] = "config://"
	cfg[base.ConfigAccessToken] = hosted.ConfigAccessToken
	cfg[base.ConfigType] = config.CN_INSTANCE_TYPE__HOST
	cfgs := mapToTarString(cfg)

	// -- CREATE CONTAINER
	sspc := make(map[string]interface{})
	hosted.ConfigContainerId = base.SnipId(createContainer(&hosted, cli, configDockerImage, // TODO: Hardcode to localhost for now.
		strconv.Itoa(config.RESERVED_CONFIG_SERVICE_PORT), sspc, &cspc, gctx.CommandLine, gctx.Debugger, ""))

	// INJECT DEBUGGER FLAG
	if gctx.Debugger > 0 {
		dt, err := manage.StringToTar(base.DebuggerFlagFileName, strconv.Itoa(gctx.Debugger))
		if err != nil {
			panic("Bad debugger port.")
		}
		injectFile(cli, hosted.ConfigContainerId, "/", dt)
	}

	// -- INJECT CONFIG
	injectFile(cli, hosted.ConfigContainerId, "/", cfgs)

	// -- START CONTAINER
	if err = cli.ContainerStart(dctx, hosted.ConfigContainerId, dcontainer.StartOptions{}); err != nil {
		panic(base.NewGinfraErrorChild("Could not start container.", err))
	}

	// Check the container.
	// TODO: probably don't want to do this here since we really don't know the address yet.
	//if _, err = ifx.PingService(gctx.PortalHostnameUrl, strconv.Itoa(config.RESERVED_CONFIG_SERVICE_PORT),
	//	app.ConfigServiceName, "Configuration provider"); err != nil {
	//	panic(base.NewGinfraErrorChild("Configuration provider did not start.", err))
	//}

	// Get container information.
	cjson = InspectContainer(cli, hosted.ConfigContainerId)
	hosted.ConfigUrl = fmt.Sprintf("http://%s:%d", cjson.NetworkSettings.Networks[hosted.NetworkName].IPAddress,
		config.RESERVED_CONFIG_SERVICE_PORT)

	// If host is set make i the primary host.
	if err == nil && gctx.PortalHostnameUrl != "" {
		hosted.ConfigPortalHostUrl = fmt.Sprintf("http://%s:%d", gctx.PortalHostnameUrl,
			config.RESERVED_CONFIG_SERVICE_PORT)
	} else {
		hosted.ConfigPortalHostUrl = hosted.ConfigUrl
	}

	if err == nil {
		// Write out file.
		if err = common.WriteJsonFile(hosted, manage.GetHostFilePath(gctx),
			"hosting metadata"); err != nil {
			panic(base.NewGinfraErrorChildA("Could not write hosting metadata file.", err, base.LM_FILEPATH,
				app.PathHostingFilePath))
		}
		gctx.GetLogger().Info("HostFile deployment initialized.  Configuration service alive.", base.LM_SERVICE_IP,
			hosted.ConfigUrl, base.LM_SERVICE_PORT, config.RESERVED_CONFIG_SERVICE_PORT)
	}

	return err
}

func HostStart(gctx *app.GContext, configContainerImage string, additional []string) (err error) {

	var (
		hosted manage.HostFile
		cprv   config.ConfigProvider
		stoken string
		cid    string
		cjson  dtypes.ContainerJSON
		cspc   *shared.Specifics      // Container specifics.
		sspc   map[string]interface{} // Service specifics.
	)

	// -- Get deployment info and config provider. -------------------------------------------------------------
	if hosted, cprv, err = prepHosted(gctx); err != nil {
		return err
	}

	// -- Get specifics ----------------------------------------------------------------------------------------
	if sspc, err = manage.GetServiceSpecifics(gctx, additional); err != nil {
		return base.NewGinfraErrorChild("Could not process specifics for service.", err)
	}

	cli, err := dockerc.NewClientWithOpts(dockerc.FromEnv, dockerc.WithAPIVersionNegotiation())
	if err != nil {
		panic(fmt.Sprintf("Could not get docker client: %v", err))
	}
	defer func() {
		_ = cli.Close() // Little late to do anything about it.
	}()

	if cspc, err = getContainerSpecifics(cli, configContainerImage, sspc); err != nil {
		// It's ok to not have specifics.
		if ((base.GetErrorFlags(err) & base.ERRFLAG_NOT_FOUND) > 0) ||
			((base.GetErrorFlags(err) & base.ERRFLAG_RECOVERABLE) > 0) {
			gctx.GetLogger().Warn("Bad specifics for container.  Ignoring.", zap.String(base.LM_CID, cid), zap.Error(err))
		} else {
			return err
		}
	}

	if gctx.Port <= 0 {
		gctx.Port = getServicePort(gctx, cprv, &hosted, cspc.Meta.Name)
	}

	if cspc, err = manage.PrepSpecifics(gctx, cspc, sspc); err != nil {
		return err
	}

	// -- Deal with security specifics -------------------------------------------------------------------------------

	if gctx.Bundle != "" {
		if ok, cert, key, errr := shared.BundleExtractServerTLS(base.GetGinfraHome()); ok {
			sspc[config.CRD_SSERVICE__CONFIG__SPECIFICS__SRVCERT] = cert
			sspc[config.CRD_SSERVICE__CONFIG__SPECIFICS__SRVKEY] = key
		} else if errr != nil {
			gctx.GetLogger().Warnw("Bundle configured but could not be extracted.", zap.Error(errr))
		}
	}

	// -- Register service. --------------------------------------------------------------------------------------

	// Service class
	sclass := gctx.ServiceClass
	if sclass == "" {
		sclass = cspc.Meta.Name
	}

	auth := gctx.Authtoken
	if auth == "" {
		auth = base.GenerateCodeword(base.ConfigAuthLength)
		gctx.GetLogger().Info("Using generated auth token.")
	} else {
		gctx.GetLogger().Info("Using provided auth token.")
	}
	if stoken, err = cprv.RegisterService(config.ITHost, cspc.Meta.Name, hosted.DeploymentName, auth, configContainerImage, hosted.ConfigPortalHostUrl,
		sclass, gctx.Port, sspc); err != nil {
		return base.NewGinfraErrorChild("Could not register service", err)
	}

	// From here rollbacks become necessary.
	// -- Docker client ---------------------------------------------------------------------------------------

	// Make sure we rollback even if there is a panic from something else.
	defer func() {
		if r := recover(); r != nil {
			if perr, ok := r.(*base.GinfraError); ok {
				err = perr
			} else {
				err = base.NewGinfraErrorA("Failed to start service.", base.LM_CAUSE, fmt.Sprintf("%v", r))
			}
			manage.RollbackRegistration(gctx, hosted.DeploymentName, cspc.Meta.Name, stoken, cprv)
			if !gctx.NoRollback {
				rollbackContainer(gctx, cid, cli)
			}

			// TODO reap services no longer needed.
		}
	}()

	// -- CREATE CONFIG
	var cfg = make(map[string]interface{})
	cfg[base.ConfigAuth] = auth
	cfg[base.ConfigUri] = hosted.ConfigUrl
	cfg[base.ConfigPortalUri] = hosted.ConfigPortalHostUrl
	cfg[base.ConfigAccessToken] = stoken
	cfg[base.ConfigDeployment] = gctx.Deployment
	cfg[base.ConfigName] = cspc.Meta.Name
	cfg[base.ConfigPort] = gctx.Port
	cfg[base.ConfigType] = config.CN_INSTANCE_TYPE__HOST
	if gctx.Clear {
		cfg[base.ConfigClear] = "true"
	}
	cfg[base.ConfigSpecifics] = sspc

	// -- CREATE CONTAINER
	cid = createContainer(&hosted, cli, configContainerImage, strconv.Itoa(gctx.Port), sspc, cspc, gctx.CommandLine,
		gctx.Debugger, gctx.Gpus)
	cid = base.SnipId(cid)
	cfg[base.ConfigId] = cid
	cfgs := mapToTarString(cfg)

	// INJECT DEBUGGER FLAG
	if gctx.Debugger > 0 {
		dt, err := manage.StringToTar(base.DebuggerFlagFileName, strconv.Itoa(gctx.Debugger))
		if err != nil {
			panic("Bad debugger port.")
		}
		injectFile(cli, cid, "/", dt)
	}

	// INJECT NOSERVICE FLAG for DontStart
	if gctx.DontStart {
		dt, _ := manage.StringToTar(base.NoServiceFlagFileName, "true")
		injectFile(cli, cid, "/", dt)
	}

	// -- INJECT CONFIG FILE
	injectFile(cli, cid, "/", cfgs)

	// -- START CONTAINER
	if err = cli.ContainerStart(context.Background(), cid, dcontainer.StartOptions{}); err != nil {
		panic(base.NewGinfraErrorChild("Could not start container.", err))
	}

	// -- REGISTER NEW CONTAINER
	cjson = InspectContainer(cli, cid)
	fqdn := cjson.NetworkSettings.Networks[hosted.NetworkName].IPAddress // Will make this real later.
	if stoken, err = cprv.RegisterInstance(config.ITHost, cspc.Meta.Name, hosted.DeploymentName, stoken, cid,
		cjson.NetworkSettings.Networks[hosted.NetworkName].IPAddress,
		fqdn); err != nil { // TODO Look into using hostname instead.
		panic(base.NewGinfraErrorChild("Could not register service instance.", err))
	}

	// -- HANDLE NEW CONTAINER
	cjson = HandleFreshContainer(cli, cid, strconv.Itoa(gctx.Port), hosted.NetworkName)

	if err == nil {

		// Add discovery for primary port.
		err = cprv.DiscoveryAdd(gctx.Deployment, sclass, fmt.Sprintf("http://%s:%d/", fqdn, gctx.Port))
		if err != nil {
			gctx.GetLogger().Warn("Could not add discovery for primary port.", zap.Error(err))
		}

		// Add discovery for exports.
		if cspc != nil {
			for nport, e := range cspc.Exposed {

				if nport != "" && e.ServiceClass.Name != "" && e.ServiceClass.Url != "" {
					nurl, err := common.SimpleReplacer(e.ServiceClass.Url,
						map[string]interface{}{"host": fqdn, "port": nport}, false)
					if err != nil {
						gctx.GetLogger().Panic("Container specifics are corrupt.", zap.Error(err))
					}

					err = cprv.DiscoveryAdd(gctx.Deployment, cspc.Meta.Name+"/"+e.ServiceClass.Name, nurl)
					if err != nil {
						gctx.GetLogger().Warn("Could not add discovery for exposed port.", zap.Error(err))
					}

				}

			}
		}

		// If host is set make it the primary host.  This is a little heavy-handed since it happens every time, but
		// hosting like this aren't very big.
		_, err = cprv.PortalAdd(hosted.DeploymentName, config.CO_ANY_PORTAL, hosted.DockerHost)
		if err != nil {
			panic(err)
		}

		gctx.GetLogger().Infow("HostFile service started.", base.LM_SERVICE_IP,
			cjson.NetworkSettings.Networks[hosted.NetworkName].IPAddress, base.LM_SERVICE_PORT, gctx.Port)
	}

	return err
}
