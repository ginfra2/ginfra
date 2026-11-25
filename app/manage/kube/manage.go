/*
Package kube
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes hosted deployments.
*/
package kube

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

const ()

// #####################################################################################################################
// #

func KubePortal(gctx *app.GContext, name, addy string) error {

	var ()

	// -- Get deployment info and config provider.
	kube, cprv, err := prepKube(gctx)
	if err != nil {
		return err
	}

	if name == "" || name == "default" {
		name = config.CO_ANY_PORTAL
	}

	_, err = cprv.PortalAdd(kube.DeploymentName, name, addy)
	if err != nil {
		return err
	}

	return nil
}

func KubeInitData(gctx *app.GContext, name, initName string) (err error) {
	// -- Get deployment info and config provider.
	kube, cprv, err := prepKube(gctx)
	if err != nil {
		return err
	}

	s := config.GetServiceConfigN(cprv, kube.DeploymentName, name)
	if s.Err != nil {
		return base.NewGinfraErrorChild("Service and deployment not registered.", s.Err)
	}

	var sclass string
	sclass, err = s.GetValue(config.CRD_SSERVICE__CONFIG__SERVICE_CLASS).GetValueString()
	if err != nil {
		return base.NewGinfraErrorA("Service class not set for service.", base.LM_SERVICE_NAME, name)
	}

	var prt string
	prt, err = cprv.DiscoverPortal(kube.DeploymentName, sclass)
	if err != nil {
		return base.NewGinfraErrorA("Portal not exposed for service.", base.LM_SERVICE_NAME, name)
	}

	auth := config.GetServiceAuthN(cprv, kube.DeploymentName, name)
	if s.Err != nil {
		gctx.GetLogger().Warn("No auth set for service, so it can't be init'd.",
			zap.String(base.LM_SERVICE_NAME, name))

	} else {
		// Order shutdown
		var lc *control.ClientWithResponses
		lc, err = control.NewClientWithResponses(prt + "/" + name)

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
					kube.DeploymentName, name))
			}

			if r.StatusCode() == http.StatusOK {
				gctx.GetLogger().Infow("Init succeeded.", zap.String("Uri", r.JSON200.Uri),
					zap.Int("Port", r.JSON200.Port), zap.String("Username", r.JSON200.Username),
					zap.String("Password", r.JSON200.Password))
			} else {
				gctx.GetLogger().Errorw("Failed to init service/server.", zap.Int(base.LM_CODE, r.StatusCode()),
					zap.String(base.LM_CAUSE, string(r.Body)))
			}
		}

		if err != nil {
			gctx.GetLogger().Warn("Could not init service.", zap.String(base.LM_SERVICE_NAME, name))

		}

	}

	return err
}
