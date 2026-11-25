/*
Package local
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Start local processes.
*/
package local

import (
	"fmt"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// #####################################################################################################################
// #

func addDiscoveryForExports(gctx *app.GContext, cprv config.ConfigProvider, ex map[string]shared.Expose) {
	for nport, e := range ex {

		if nport != "" && e.ServiceClass.Name != "" && e.ServiceClass.Url != "" {
			nurl, err := common.SimpleReplacer(e.ServiceClass.Url,
				map[string]interface{}{"host": "localhost", "port": nport}, false)
			if err != nil {
				gctx.GetLogger().Panic("Container specifics are corrupt.", zap.Error(err))
			}

			err = cprv.DiscoveryAdd(gctx.Deployment, e.ServiceClass.Name, nurl)
			if err != nil {
				gctx.GetLogger().Warn("Could not add discovery for exposed port.", zap.Error(err))
			}

		}
	}
}

// StartLocal initializes and starts a local service using the specified context, configuration provider URI, and additional options.
// It sets up the service environment, reads and validates required files, registers the service, and handles optional parameters.
// If the service is configured to start, it executes the appropriate start command. Returns an error if any step fails.
func StartLocal(gctx *app.GContext, configProviderUri string, additional []string) error {

	var (
		err    error
		cwd    string
		cprv   config.ConfigProvider
		stoken = config.CONFIG_NIL_AUTH
	)

	// Get current wd
	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	// Get local working dir if path not set.
	if gctx.ServiceHomePath == "" {
		// Use local working dir as service path
		gctx.ServiceHomePath = cwd
	}

	// Make sure we are in the service directory.
	err = os.Chdir(gctx.ServiceHomePath)
	if err != nil {
		panic(err)
	}

	// Load container specifics.
	cspec, err := shared.SpecificsLoadFile(filepath.Join(gctx.ServiceHomePath, base.SpecificsFileName), map[string]interface{}{})
	if err != nil {
		return base.NewGinfraErrorChild("Specifics (container) could not be loaded from the service.  Every service must have a specifics.json file in the service root directory.",
			err)
	}
	if cspec == nil {
		panic(err)
	}

	rcp := config.IfFileResolvePath(configProviderUri, gctx.ServiceHomePath)
	if cprv, err = config.GetConfigProvider(rcp, stoken); err != nil {
		return base.NewGinfraErrorChildA("Could not load configuration provider", err, base.LM_CONFIG_PATH,
			rcp)
	}

	// -- CREATE CONFIG FILE.
	var cfg = make(map[string]interface{})
	cfg[base.ConfigAuth] = config.CONFIG_NIL_AUTH
	cfg[base.ConfigUri] = configProviderUri
	cfg[base.ConfigPortalUri] = configProviderUri
	cfg[base.ConfigAccessToken] = stoken
	cfg[base.ConfigDeployment] = gctx.Deployment
	cfg[base.ConfigName] = cspec.Meta.Name
	cfg[base.ConfigPort] = gctx.Port
	cfg[base.ConfigType] = config.CN_INSTANCE_TYPE__LOCAL
	cfg[base.ConfigId] = config.CN_INSTANCE_TYPE__LOCAL

	if gctx.Clear {
		cfg[base.ConfigClear] = "true"
	}

	// Register only configurable services.
	if cprv.IsConfigurable() {

		// -- REGISTER SERVICE --------------------------------------------------------------------
		//RegisterService(name, deployment, image, configProvider string, port int32, sspecs *[]models.Namevalue)

		var sspecs map[string]interface{}
		if sspecs, err = manage.GetServiceSpecifics(gctx, additional); err != nil {
			return base.NewGinfraErrorChild("Could not process sspecs for service.", err)
		}
		cfg[base.ConfigSpecifics] = sspecs

		if stoken, err = cprv.RegisterService(config.ITLocal, cspec.Meta.Name, gctx.Deployment, config.CONFIG_NIL_AUTH,
			"", rcp, cspec.Meta.ServiceClass, gctx.Port, sspecs); err != nil {
			return base.NewGinfraErrorChild("Could not register service", err)
		}

		// -- REGISTER SERVICE INSTANCE -----------------------------------------------------------
		// Usually this would happen after start but since this is running on the host, we already know what we need.
		if stoken, err = cprv.RegisterInstance(config.ITLocal, cspec.Meta.Name, gctx.Deployment, stoken, config.CN_INSTANCE_TYPE__LOCAL,
			"127.0.0.1", "localhost"); err != nil {
			return base.NewGinfraErrorChild("Could not register service instance.", err)
			// TODO reap services no longer needed.
		}

		// -- ADD DISCOVER ---------------------------------------------------------------------
		err = cprv.DiscoveryAdd(gctx.Deployment, cspec.Meta.ServiceClass, fmt.Sprintf("http://localhost:%d/", gctx.Port))
		if err != nil {
			gctx.GetLogger().Warn("Could not add discovery for primary port.", zap.Error(err))
		}

		// Add discovery for exports.
		addDiscoveryForExports(gctx, cprv, cspec.Exposed)

	}

	// -- SET CONFIG -----------------------------------------------------------------------
	if gctx.Clear == true {
		cfg[base.ConfigClear] = "true"
	}

	if err = common.WriteJsonFileA(cfg, path.Join(gctx.ServiceHomePath, base.ConfigFileName),
		"configuration file "+base.ConfigFileName); err != nil {
		return base.NewGinfraErrorChildA("Could not write configuration file.", err, base.LM_FILEPATH,
			path.Join(gctx.ServiceHomePath, base.ConfigFileName))
	}

	// -- START SERVICE ---------------------------------------------------------------------
	if gctx.DontStart {
		gctx.GetLogger().Infow("Service not started as ordered.")

	} else {
		var cmd *exec.Cmd
		pm := config.CommandBuilder(gctx.ServiceHomePath) // Greatly simplified in the latest version.
		if cspec.Meta.InvocationCmd == "" {
			if runtime.GOOS == "windows" {
				cspec.Meta.InvocationCmd = "start_local.cmd"
			} else {
				cspec.Meta.InvocationCmd = gctx.ServiceHomePath + "/start_local.sh"
			}
		}
		if runtime.GOOS == "windows" {
			cspec.Meta.InvocationCmd = strings.Replace(cspec.Meta.InvocationCmd, "/", "\\", -1)
			pm = append([]string{"/c", cspec.Meta.InvocationCmd}, pm...)
			cmd = exec.Command("cmd.exe", pm...)
		} else {
			cmd = exec.Command(cspec.Meta.InvocationCmd, pm...)
		}
		cmd.Dir = gctx.ServiceHomePath

		err = app.RunCommandInContextLive(gctx, cmd, cspec.Meta.Name)
		if err == nil {
			gctx.GetLogger().Infow("Started service.  Kill it to shut it down.", base.LM_SERVICE_NAME,
				cspec.Meta.Name, base.LM_COMMAND, cmd.String())
		}

	}
	return err
}
