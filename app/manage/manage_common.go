/*
Package manage
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Common tools and utilities.

The host file describes a deployment and how to access it.  Typically the commands will find this in the CWD, but there
are parameters to override that.  Each command will work with only one deployment at a time, thus only one of these
files at a time.  While technically the file is mutable, it should generally be left alone after initialization, which
it accomplished by one of the Init commands.
*/
package manage

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
)

// Hostfile

type HostFile struct {
	DeploymentName string `json:"DeploymentName"`

	ConfigAccessToken   string `json:"ConfigAccessToken"`
	ConfigUrl           string `json:"ConfigUrl"`
	ConfigPortalHostUrl string `json:"ConfigPortalHostUrl"`

	NextServicePort int `json:"NextServicePort"`

	// Host specific
	ConfigContainerId string `json:"ConfigContainerId"`
	DockerHost        string `json:"DockerHost"`
	NetworkName       string `json:"NetworkName"`
	NetworkId         string `json:"NetworkId"`

	// Kube specific
	InitConfigMapName string `json:"InitConfigMapName"`
}

func GetServiceSpecifics(gctx *app.GContext, additional []string) (map[string]interface{}, error) {
	var (
		err error
	)
	specifics := make(map[string]interface{}) // Can't be nil because the API will freak out.

	if gctx.Specifics != "" {
		if specifics, err = common.StringToJsonmap(gctx.Specifics); err != nil {
			return nil, base.NewGinfraErrorChildA("Bad value for specifics (--specifics json).", err, base.LM_TEXT,
				gctx.Specifics)
		}
	}
	if len(additional) > 0 {

		// Additional specifics.
		if additional != nil && len(additional) > 0 {

			if (len(additional) % 2) == 0 {
				for i := 0; i < len(additional); i = i + 2 {
					specifics[additional[i]] = additional[i+1]
				}
			} else {
				return nil, base.NewGinfraErrorA("Bad value for specifics (command line).  A name is not paired with a value.")
			}

		}
	}
	return specifics, err
}

func GetHostFilePath(gctx *app.GContext) string {
	return path.Join(gctx.HostDirPath, app.PathHostingFilePath)
}

func LoadHostFileCwd() (*HostFile, error) {
	var hf HostFile

	cwd, err := os.Getwd()
	if err == nil {
		if _, err = os.Stat(filepath.Join(cwd, app.PathHostingFilePath)); err == nil {
			// File exists
			hf, err = common.ReadJsonFile[HostFile](filepath.Join(cwd, app.PathHostingFilePath), "hosting metadata")
		}
	}

	if err != nil {
		return nil, base.NewGinfraErrorChild("Hostfile.json could not be loaded.", err)
	}

	return &hf, nil
}

func GetConfigProvider(url, accessToken string) (config.ConfigProvider, error) {
	var (
		prov config.ConfigProvider
		err  error
	)
	if prov, err = config.GetConfigProvider(url, accessToken); err != nil {
		err = base.NewGinfraErrorChildA("Could not load configuration provider", err, base.LM_CONFIG_PATH,
			url)
	}
	return prov, err
}

func RemoveRegistration(deployment, name string, cprv config.ConfigProvider) error {
	return cprv.DeregisterService(name, deployment)
}

func RollbackRegistration(gctx *app.GContext, deployment, name, stoken string, cprv config.ConfigProvider) {
	if stoken != "" {
		gctx.GetLogger().Info("Rolling back registration.")
		if err := RemoveRegistration(deployment, name, cprv); err != nil {
			gctx.GetLogger().Errorw("Failed to deregister service during rollback.",
				zap.String(base.LM_CAUSE, err.Error()))
		}
	}
}

func PrepSpecifics(gctx *app.GContext, cspec *shared.Specifics, sspec map[string]interface{}) (*shared.Specifics, error) {
	var (
		out *shared.Specifics
		err error
	)
	repsspc := sspec

	repsspc[base.ConfigPort] = gctx.Port // Override port
	if _, ok := repsspc[base.ConfigWebPort]; !ok {
		repsspc[base.ConfigWebPort] = base.DefaultWebPort
	}

	if out, err = shared.SpecificsProcess(cspec, repsspc); err != nil {
		return nil, err
	}
	return out, nil
}
