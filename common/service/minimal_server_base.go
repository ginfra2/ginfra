/*
Package service
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Minimal service.
*/
package service

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"gitlab.com/ginfra/ginfra/api/service_control/models"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

// ===============================================================================================
// SERVER IMPLEMENTATION

type MinimalServer struct {
	EchoServer *echo.Echo
	homepath   string
	Err        error
	ErrorState bool
	port       int
}

func InitServer(port int) (*MinimalServer, error) {

	sc := &MinimalServer{}
	sc.port = port
	sc.EchoServer = echo.New()
	sc.EchoServer.HideBanner = true
	sc.EchoServer.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodHead},
	}))

	return sc, nil
}

func RunServer(sc *MinimalServer) error {

	go func(scc *MinimalServer) {
		// I may have to morph sc later.
		scc.Err = scc.EchoServer.Start(fmt.Sprintf("%v:%v", "localhost", sc.port))
		if scc.Err != nil && scc.Err.Error() != "http: Server closed" {
			app.GetRootContext().GetLogger().Error("Server exited with an error", zap.Error(scc.Err))
		}
	}(sc)

	timeout := time.After(2 * time.Second)
	tick := time.Tick(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return base.NewGinfraError("Server did not start in time.")
		case <-tick:
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", 44244), 100*time.Millisecond)
			if err == nil {
				conn.Close()
				goto PortOpen
			}
		}
	}
PortOpen:
	return nil
}

func StopServer(sc *MinimalServer) error {
	if sc.EchoServer == nil {
		panic("Service not running.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := sc.EchoServer.Shutdown(ctx)

	sc.EchoServer = nil
	return err
}

func (ms *MinimalServer) SetErrorState(state bool) {
	ms.ErrorState = state
}

// =======================================================================================
// = SERVICE IMPLEMENTATIONS

func GetErrorResponse(err error) models.ErrorResponse {
	m := err.Error()
	e := models.ErrorResponse{
		Message: &m,
	}
	return e
}

func PostErrorResponse(ctx echo.Context, code int, err error) error {
	err = ctx.JSON(code, GetErrorResponse(err))
	return err
}

func PostErrorUnauthorizedResponse(ctx echo.Context, path string, err error) error {
	err = ctx.JSON(http.StatusUnauthorized, GetErrorResponse(base.NewGinfraErrorChildA("Not authorized.", err,
		base.LM_CONFIG_PATH, path)))
	return err
}

func PostErrorNotFoundResponse(ctx echo.Context, err error, target string) error {
	err = ctx.JSON(http.StatusNotFound, GetErrorResponse(
		base.NewGinfraErrorChildA("Not found.", err, base.LM_TARGET, target)))
	return err
}
