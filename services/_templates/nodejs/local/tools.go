/*
Package local
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt if unaltered.  You are welcome to alter and apply any
license you wish to this file.

Tools.
*/
package local

import (
	"<<<Module>>>/api/service/models" // If you use multiple version, this will have to be changed.
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/config"
	"go.uber.org/zap"
	"net/http"
)

func IsAuthorized(auth string, auths []string) bool {
	for _, i := range auths {
		if i == auth {
			return true
		}
	}
	return false
}

func GetErrorResponse(err error) models.ErrorResponse {
	m := err.Error()
	e := models.ErrorResponse{
		Message: &m,
	}
	return e
}

func PostErrorResponse(ctx echo.Context, code int, err error) error {
	if errr := ctx.JSON(code, GetErrorResponse(err)); errr != nil {
		Logger.Warn("Problem returning ERROR to client.", zap.Error(errr))
	}
	return err
}

func PostErrorUnauthorizedResponse(ctx echo.Context, path string, err error) error {
	if errr := ctx.JSON(http.StatusUnauthorized, GetErrorResponse(base.NewGinfraErrorChildA("Not authorized.", err,
		base.LM_CONFIG_PATH, path))); errr != nil {
		Logger.Warn("Problem returning UNAUTHORIZED to client.", zap.Error(errr))
	}
	return err
}

func PostErrorNotFoundResponse(ctx echo.Context, err error, target string) error {
	if errr := ctx.JSON(http.StatusNotFound, GetErrorResponse(
		base.NewGinfraErrorChildA("Not found.", err, base.LM_TARGET, target))); errr != nil {
		Logger.Warn("Problem returning NOT FOUND to client.", zap.Error(errr))
	}
	return err
}

func CheckAndPostTerminalOk(ctx echo.Context, err error) error {
	if err != nil {
		panic(fmt.Sprintf("BUG: Escaped Error.  %v", err))
	}
	if errr := ctx.String(http.StatusOK, "OK"); errr != nil {
		Logger.Warn("Problem returning OK to client.", zap.Error(errr))
	}
	return nil
}

func GetConfiguredServiceSpecifics(gtcx *GContext) *config.Config {
	scfg := config.GetServiceConfig(gtcx.ConfigProvider, gtcx.Scfg.DeploymentName)
	if scfg.Err == nil {
		scfg = scfg.GetConfig(config.CRD_SSERVICE__CONFIG__SPECIFICS)
	}
	if scfg.Err != nil {
		scfg.Err = base.NewGinfraErrorChild("Could not get configured service specifics from the configuration provider.", scfg.Err)
	}
	return scfg
}

func ProcessError(ctx echo.Context, err error, path string) error {
	Logger.Error("Error response.", zap.String(base.LM_CAUSE, err.Error()))
	var cerr *base.GinfraError
	if errors.As(err, &cerr) {
		if (cerr.GetFlags() & base.ERRFLAG_NOT_AUTHORIZED) > 0 {
			return PostErrorUnauthorizedResponse(ctx, path, cerr)
		} else if (cerr.GetFlags() & base.ERRFLAG_NOT_FOUND) > 0 {
			return PostErrorNotFoundResponse(ctx, cerr, path)
		} else if (cerr.GetFlags() & base.ERRFLAG_SUBSYSTEM_SERIOUS) > 0 {
			return PostErrorResponse(ctx, http.StatusInternalServerError, err)
		} else {
			return PostErrorResponse(ctx, http.StatusBadRequest, cerr)
		}
	} else {
		return PostErrorResponse(ctx, http.StatusInternalServerError, err)
	}
}
