/*
Package local
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Minimal service.
*/
package local

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"gitlab.com/ginfra/ginfra/api/service_control/models"
	"gitlab.com/ginfra/ginfra/api/service_control/server"
	"gitlab.com/ginfra/ginfra/base"
	"net/http"
	"time"
)

// ===============================================================================================
// SERVER IMPLEMENTATION

type MinimalLocalService struct {
	echoServer *echo.Echo
	homepath   string
	Err        error
	errorState bool
}

func RunService(port int) (*MinimalLocalService, error) {

	sc := &MinimalLocalService{}
	sc.echoServer = echo.New()
	sc.echoServer.HideBanner = true
	sc.echoServer.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodHead},
	}))

	server.RegisterHandlers(sc.echoServer, sc)

	go func(scc *MinimalLocalService) {
		// I may have to morph sc later.
		scc.Err = scc.echoServer.Start(fmt.Sprintf("%v:%v", "localhost", port))
	}(sc)

	return sc, nil
}

func StopService(sc *MinimalLocalService) error {
	if sc.echoServer == nil {
		panic("Service not running.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := sc.echoServer.Shutdown(ctx)

	sc.echoServer = nil
	return err
}

func (ms *MinimalLocalService) SetErrorState(state bool) {
	ms.errorState = state
}

// =======================================================================================
// = SERVICE IMPLEMENTATIONS

func getErrorResponse(err error) models.ErrorResponse {
	m := err.Error()
	e := models.ErrorResponse{
		Message: &m,
	}
	return e
}

func postErrorResponse(ctx echo.Context, code int, err error) error {
	err = ctx.JSON(code, getErrorResponse(err))
	return err
}

func postErrorUnauthorizedResponse(ctx echo.Context, path string, err error) error {
	err = ctx.JSON(http.StatusUnauthorized, getErrorResponse(base.NewGinfraErrorChildA("Not authorized.", err,
		base.LM_CONFIG_PATH, path)))
	return err
}

func postErrorNotFoundResponse(ctx echo.Context, err error, target string) error {
	err = ctx.JSON(http.StatusNotFound, getErrorResponse(
		base.NewGinfraErrorChildA("Not found.", err, base.LM_TARGET, target)))
	return err
}

func checkAndPostTerminalOk(ctx echo.Context, err error) error {
	if err != nil {
		panic(fmt.Sprintf("BUG: Escaped Error.  %v", err))
	}
	err = ctx.String(http.StatusOK, "OK")
	return err
}

// GiControlDatasourceInit Initialize a data source.
// (POST /ginfra/datasource/init)
func (ms *MinimalLocalService) GiControlDatasourceInit(ctx echo.Context, params models.GiControlDatasourceInitParams) error {
	if ms.errorState {
		return postErrorResponse(ctx, http.StatusInternalServerError, base.NewGinfraError("Service is in error state."))
	}
	var r models.ControlDatasourceInitResponse
	r.Uri = "http://localhost:44244/"
	r.Port = 44244
	r.Username = "admin"
	r.Password = "<PASSWORD>"
	return ctx.JSON(http.StatusOK, r)
}

// GiControlManageReset Reset the service/server.
// (POST /ginfra/manage/reset)
func (ms *MinimalLocalService) GiControlManageReset(ctx echo.Context, params models.GiControlManageResetParams) error {
	return ctx.JSON(http.StatusOK, "{}")
}

// GiControlManageStop Stop the service/server.
// (POST /ginfra/manage/stop)
func (ms *MinimalLocalService) GiControlManageStop(ctx echo.Context, params models.GiControlManageStopParams) error {
	return ctx.JSON(http.StatusOK, "{}")
}

// GiGetSpecifics Get specifics for this service
// (GET /ginfra/specifics)
func (ms *MinimalLocalService) GiGetSpecifics(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "{}")
}
