/*
Package local
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Other local processes.
*/
package local

import (
	"context"
	"fmt"
	control "gitlab.com/ginfra/ginfra/api/service_control/client"
	cmodel "gitlab.com/ginfra/ginfra/api/service_control/models"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/config"
	"go.uber.org/zap"
	"net/http"
)

// #####################################################################################################################
// #

// LocalInitData initializes a local data service by resolving configuration, obtaining authentication, and invoking APIs.
// It uses the provided service context, configuration URI, service name, and initialization name.
// Returns an error if initialization fails at any step.
func LocalInitData(gctx *app.GContext, configProviderUri string, name, initName string) (err error) {

	rcp := config.IfFileResolvePath(configProviderUri, gctx.ServiceHomePath)
	cprv, err := config.GetConfigProvider(rcp, "A")
	if err != nil {
		return base.NewGinfraErrorChildA("Could not load configuration provider", err, base.LM_CONFIG_PATH,
			rcp)
	}

	s := config.GetServiceConfigN(cprv, config.DEFAULT_DEPLOYMENT_NAME, name)
	if s.Err != nil {
		return base.NewGinfraErrorChild("Service and deployment not registered.", s.Err)
	}

	auth := config.GetServiceAuthN(cprv, config.DEFAULT_DEPLOYMENT_NAME, name)
	if auth.GetValueError() != nil {
		gctx.GetLogger().Warn("No auth set for service, so it can't be init'd.",
			zap.String(base.LM_SERVICE_NAME, name))

	} else {
		var lc *control.ClientWithResponses
		lc, err = control.NewClientWithResponses("http://localhost" +
			":" + s.GetValue(config.CL_PORT).GetValueStringNoerr())

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
			if err != nil {
				err = base.NewGinfraErrorChild("Failed to init service/server due to service or api problem.", err)
			} else {

				if r == nil {
					panic(fmt.Sprintf("whoops  cannot have both a nil response and no error: %s / %s",
						config.DEFAULT_DEPLOYMENT_NAME, name))
				}

				if r.StatusCode() == http.StatusOK {
					gctx.GetLogger().Infow("Init succeeded.", zap.String("Uri", r.JSON200.Uri),
						zap.Int("Port", r.JSON200.Port), zap.String("Username", r.JSON200.Username),
						zap.String("Password", r.JSON200.Password))
				} else {
					err = base.NewGinfraErrorA("Failed to init service/server.", base.LM_CODE, r.StatusCode(),
						base.LM_CAUSE, string(r.Body))
				}
			}

		} else {
			gctx.GetLogger().Warn("Could not create client.", zap.String(base.LM_SERVICE_NAME, name))

		}

	}

	return err
}
