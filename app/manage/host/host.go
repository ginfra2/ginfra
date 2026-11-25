/*
Package host
Copyright (c) 2023, 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage hosted deployments which are either local docker container host.  All the commands in this file rely on
host initialization having occurred using the host.HostInit command.  They will look for the control metadata file
for the hosted deployment, which will either be in the CWD or pointed at through another command parameter (outside of the
scope of the code in this file).  A single container host can support multiple deployments, but a single command
invocation will deal with only one of them at a time.

TODO: support external container runtimes.  Right now it is hardcoded to look for a local runtime only.
*/
package host

import (
	"context"
	"encoding/json"
	"fmt"
	dockerc "github.com/docker/docker/client"
	control "gitlab.com/ginfra/ginfra/api/service_control/client"
	cmodel "gitlab.com/ginfra/ginfra/api/service_control/models"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"go.uber.org/zap"
	"net/http"
	"os"
)

// #####################################################################################################################
// #

// prepHosted returns the populated HostFile object, ConfigProvider object, ConfigProvider URI, and error.
// It reads hosting metadata from a JSON file and fetches the ConfigProvider based on the provided information.
// Input:
// - gctx: The GContext object containing deployment information and logger.
// Output:
// - hosted: The HostFile object with deployment information.
// - prov: The ConfigProvider object.
// - provuri: The URI of the ConfigProvider.
// - err: The error, if any.
func prepHosted(gctx *app.GContext) (hosted manage.HostFile, prov config.ConfigProvider, err error) {
	// -- Get deployment info.
	hosted, err = common.ReadJsonFile[manage.HostFile](manage.GetHostFilePath(gctx), "hosting metadata")
	if err != nil {
		err = base.NewGinfraErrorChildA(`Could not load hosting metadata file (Hosting.json).  Was the deployment 
initialized, is the current working directory not the directory that you initialized it in or is it corrupt?`,
			err, base.LM_FILEPATH, app.PathHostingFilePath)
	} else {
		prov, err = manage.GetConfigProvider(hosted.ConfigUrl, hosted.ConfigAccessToken)
	}
	return hosted, prov, err
}

// HostedList retrieves and lists detailed hosting information for the deployment specified in the context, which
// like all other commands will point at a single deployment metadata(/file).
// It fetches the deployment configuration and logs the deployment name, network name, and associated services.
// If no deployment exists, it logs a warning; otherwise, it outputs the relevant details or returns an error.
func HostedList(gctx *app.GContext) (err error) {

	// -- Get deployment info and config provider.
	hosted, cprv, err := prepHosted(gctx)
	if err != nil {
		return err
	}

	c := cprv.FetchConfig(config.CR_DEPLOYMENTS, hosted.DeploymentName, config.CRD_SERVICES)
	if c.Err != nil {
		if (base.GetErrorFlags(c.Err) & base.ERRFLAG_NOT_FOUND) > 0 {
			gctx.GetLogger().Warn("Nothing to list.")
		} else {
			return base.NewGinfraErrorChild("Configuration corrupt.", c.Err)
		}
	}

	fmt.Println("Deployment name : " + hosted.DeploymentName)
	fmt.Println("Network name    : " + hosted.NetworkName)

	var s []string
	for i, _ := range c.Config {
		s = append(s, i)
	}
	fmt.Print("Services    :")
	for _, i := range s {
		fmt.Print(" " + i)
	}
	fmt.Println()

	return nil
}

// HostedShow retrieves and displays the service configuration for the specified service provided and deployment from the config.
// It uses the hosting metadata to fetch the required configuration from the ConfigProvider and outputs it in JSON format.
// Returns an error if the hosting metadata file or the configuration cannot be accessed or parsed.
func HostedShow(gctx *app.GContext, name string) (err error) {

	// -- Get deployment info and config provider.
	hosted, cprv, err := prepHosted(gctx)
	if err != nil {
		return err
	}

	c := cprv.FetchConfig(config.CR_DEPLOYMENTS, hosted.DeploymentName, config.CRD_SERVICES, name)
	if c.Err != nil {
		return base.NewGinfraErrorChild("Service and deployment doesn't exist.", c.Err)
	}

	js, err := json.MarshalIndent(c.Config, "", app.JsonIndent)
	fmt.Println(string(js))
	return err
}

func hostedStop(gctx *app.GContext, cprv config.ConfigProvider, deployment, name string) (err error) {
	var cli *dockerc.Client

	s := config.GetServiceConfigN(cprv, deployment, name)
	if s.Err != nil {
		return base.NewGinfraErrorChild("Service and deployment not registered.", s.Err)
	}

	auth := config.GetServiceAuthN(cprv, deployment, name)
	if s.Err != nil {
		gctx.GetLogger().Warn("No auth set for service, so it can't be stopped.",
			zap.String(base.LM_SERVICE_NAME, name))

	} else {

		// Order shutdown
		lc, cerr := control.NewClientWithResponses("http://localhost" +
			":" + s.GetValue(config.CL_PORT).GetValueStringNoerr() + "/" + name)

		var r *control.GiControlManageStopResponse
		a := auth.GetValueStringNoerr()
		if cerr == nil {
			param := cmodel.GiControlManageStopParams{
				Auth: &a,
			}
			r, cerr = lc.GiControlManageStopWithResponse(context.Background(), &param)
			if r == nil && cerr == nil {
				panic(fmt.Sprintf("whoops  cannot have both a nil response and no error: %s / %s", deployment, name))
			}
		}

		if cerr != nil {
			gctx.GetLogger().Warn("Could not stop service.  It will die with the container.",
				zap.String(base.LM_SERVICE_NAME, name))

		}
	}

	cli, err = dockerc.NewClientWithOpts(dockerc.FromEnv, dockerc.WithAPIVersionNegotiation())
	if err != nil {
		panic(fmt.Sprintf("Could not get docker client: %v", err))
	}
	defer func() {
		_ = cli.Close() // Little late to do anything about it.
	}()

	// Kill instance
	id, c := config.GetServiceInstanceOnly(cprv, deployment, name)
	if c.Err != nil {
		return base.NewGinfraErrorChild("Corrupt configuration for instance.", c.Err)
	}
	// Errors will just be logged since there isn't much we can do about it here.
	rollbackContainer(gctx, id, cli)

	// Zap registration.
	err = manage.RemoveRegistration(deployment, name, cprv)

	return err
}

// HostedStop stops a hosted service instance by retrieving the required deployment and service information.
// It utilizes the given GContext and service name to fetch deployment details and halt the service execution.
func HostedStop(gctx *app.GContext, name string) (err error) {
	// -- Get deployment info and config provider.
	hosted, cprv, err := prepHosted(gctx)
	if err != nil {
		return err
	}
	return hostedStop(gctx, cprv, hosted.DeploymentName, name)
}

// HostedDestroy destroys a hosted deployment by removing services, configurations, networks, and associated resources.
// It utilizes the provided GContext to manage operations and handles logging and error propagation for each step.
func HostedDestroy(gctx *app.GContext) (err error) {

	// -- Get deployment info and config provider.
	hosted, cprv, err := prepHosted(gctx)
	if err != nil {
		return base.NewGinfraErrorChild("Failed to get deployment and configuration information.", err)
	}

	var (
		ok      bool
		errored bool
	)

	// Remove reservations.
	if err = cprv.RemoveConfig(config.CR_RESERVATIONS, hosted.DeploymentName); err != nil {
		gctx.GetLogger().Warn("Could not remove reservations.", zap.String(base.LM_DEPLOYMENT_NAME, hosted.DeploymentName))
	}

	// Kill services.
	svc := cprv.FetchConfig(config.CR_DEPLOYMENTS, hosted.DeploymentName, config.CRD_SERVICES)
	if svc.Err == nil {
		for k, v := range svc.Config {
			if _, ok = v.(map[string]interface{}); ok {
				// It has the deplopment for the service in question.  Zap it.
				if err = hostedStop(gctx, cprv, hosted.DeploymentName, k); err != nil {
					errored = true
					gctx.GetLogger().Errorw("Failed to remove a service for the deployment.",
						zap.String(base.LM_DEPLOYMENT_NAME, hosted.DeploymentName),
						zap.String(base.LM_CAUSE, err.Error()))
				}
			} else {
				panic("Corrupt configuration.  This should never happen.")
			}
		}
		gctx.GetLogger().Info("Done stopping services.")
	}

	// Stop configuration provider service.
	var cli *dockerc.Client
	cli, err = dockerc.NewClientWithOpts(dockerc.FromEnv, dockerc.WithAPIVersionNegotiation())
	if err != nil {
		panic(fmt.Sprintf("Could not get docker client: %v", err))
	}
	defer func() {
		_ = cli.Close() // Little late to do anything about it.
	}()
	if err = removeContainer(hosted.ConfigContainerId, cli); err != nil {
		gctx.GetLogger().Errorw("Failed to stop configuration provider service.",
			zap.String(base.LM_DEPLOYMENT_NAME, hosted.DeploymentName),
			zap.String(base.LM_CAUSE, err.Error()))
	} else {
		gctx.GetLogger().Info("Configuration provider stopped.")
	}

	rollbackNetwork(gctx, hosted.NetworkId, cli)

	// Remove Hosting.json file
	err = os.Remove(manage.GetHostFilePath(gctx))

	if err == nil && errored == true {
		err = base.NewGinfraError("Destroy not successful.")
	}
	return err
}

// HostedInitData initializes and configures a hosted service deployment with the provided service and initialization name.
// It retrieves relevant deployment information, service configuration, and authentication details, logging outcomes.
// Input: gctx for context, name for the service name, and initName for the initialization identifier.
// Returns: An error if the initialization or configuration fails.
func HostedInitData(gctx *app.GContext, name, initName string) (err error) {
	// -- Get deployment info and config provider.
	hosted, cprv, err := prepHosted(gctx)
	if err != nil {
		return err
	}

	s := config.GetServiceConfigN(cprv, hosted.DeploymentName, name)
	if s.Err != nil {
		return base.NewGinfraErrorChild("Service and deployment not registered.", s.Err)
	}

	auth := config.GetServiceAuthN(cprv, hosted.DeploymentName, name)
	if s.Err != nil {
		gctx.GetLogger().Warn("No auth set for service, so it can't be init'd.",
			zap.String(base.LM_SERVICE_NAME, name))

	} else {
		var lc *control.ClientWithResponses
		lc, err = control.NewClientWithResponses("http://localhost" +
			":" + s.GetValue(config.CL_PORT).GetValueStringNoerr() + "/" + name)

		var r *control.GiControlDatasourceInitResponse
		a := auth.GetValueStringNoerr()
		if err == nil {
			param := cmodel.GiControlDatasourceInitParams{
				Auth: &a,
			}
			body := cmodel.GiControlDatasourceInitJSONRequestBody{
				Name: initName,
			}
			r, err = lc.GiControlDatasourceInitWithResponse(context.Background(), &param, body)
			if r == nil && err == nil {
				panic(fmt.Sprintf("whoops  cannot have both a nil response and no error: %s / %s",
					hosted.DeploymentName, name))
			}

			if r.StatusCode() == http.StatusOK {
				gctx.GetLogger().Infow("Init succeeded.", zap.String("Uri", r.JSON200.Uri),
					zap.Int("Port", r.JSON200.Port), zap.String("Username", r.JSON200.Username),
					zap.String("Password", r.JSON200.Password))
			} else {
				gctx.GetLogger().Errorw("Failed to init service/server.", zap.Int(base.LM_CODE, r.StatusCode()),
					zap.String(base.LM_CAUSE, string(r.Body)))
			}

		} else {
			gctx.GetLogger().Warn("Could not talk to service/serer to init.", zap.String(base.LM_SERVICE_NAME, name))

		}

	}

	return err
}

// HostedReset resets to the service.
// Input: gctx for context, name for the service name.
// Returns: An error if the reset fails.
func HostedReset(gctx *app.GContext, name string) (err error) {
	// -- Get deployment info and config provider.
	hosted, cprv, err := prepHosted(gctx)
	if err != nil {
		return err
	}

	s := config.GetServiceConfigN(cprv, hosted.DeploymentName, name)
	if s.Err != nil {
		return base.NewGinfraErrorChild("Service and deployment not registered.", s.Err)
	}

	auth := config.GetServiceAuthN(cprv, hosted.DeploymentName, name)
	if s.Err != nil {
		gctx.GetLogger().Warn("No auth set for service, so it can't be reset.",
			zap.String(base.LM_SERVICE_NAME, name))

	} else {
		var lc *control.ClientWithResponses
		lc, err = control.NewClientWithResponses("http://localhost" +
			":" + s.GetValue(config.CL_PORT).GetValueStringNoerr() + "/" + name)

		var r *control.GiControlManageResetResponse
		a := auth.GetValueStringNoerr()
		if err == nil {
			param := cmodel.GiControlManageResetParams{
				Auth: &a,
			}
			r, err = lc.GiControlManageResetWithResponse(context.Background(), &param)
			if r == nil && err == nil {
				panic(fmt.Sprintf("whoops cannot have both a nil response and no error: %s / %s",
					hosted.DeploymentName, name))
			}

			if r.StatusCode() == http.StatusOK {
				gctx.GetLogger().Infow("Reset succeeded.", zap.String("Note", *r.JSON200.Note))
			} else {
				gctx.GetLogger().Errorw("Failed to reset service/server.", zap.Int(base.LM_CODE, r.StatusCode()),
					zap.String(base.LM_CAUSE, string(r.Body)))
			}

		} else {
			gctx.GetLogger().Warn("Could not talk to service/serer to reset.", zap.String(base.LM_SERVICE_NAME, name))

		}

	}

	return err
}
