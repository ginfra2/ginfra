/*
Package configctl
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Config management tool.
*/
package configctl

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/common/config"
)

// #####################################################################################################################
// #

func SetSpecifics(gctx *app.GContext, configurl, auth, deployment, service, name, value string) error {
	cpv, err := manage.GetConfigProvider(configurl, auth)
	if err != nil {
		return err
	}
	cfg := config.SetServiceConfigSpecificsN(cpv, deployment, service, name, value)
	return cfg.Err
}
