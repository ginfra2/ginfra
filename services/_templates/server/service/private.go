/*
Package service
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt, if unaltered.  You are free to alter and apply any license
you wish to this file.

HOWEVER, If you want ginfra to maintain and update this service, LEAVE THIS ALONE.  If you alter this file and try to
use the ginfra update command, it will wipe out any of your changes.

Private service implementation.

*/
package service

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/log"
	gecho "gitlab.com/ginfra/ginfra/common/service/echo"
	"gitlab.com/ginfra/ginfra/common/service/gotel"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"<<<Module>>>/local"
	"<<<Module>>>/providers"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"
)

// #####################################################################################################################
// # SERVER

var (
	echoServer      *echo.Echo = nil
	testServiceHome            = flag.String("test_service_home", "", "")
	shutdownwg      sync.WaitGroup
	glogger         *log.GLogger
)

func doRegistration(name string) {
	// Trap the panic.  In normal usage, a service may only be registered once, but in testing it could
	// happen multiple times.
	defer func() {
		_ = recover()
	}()
	base.RegisterService(name)
}

func ginfraHTTPErrorHandler(err error, c echo.Context) {

	if c.Response().Committed {
		return
	}

	he, ok := err.(*echo.HTTPError)
	if ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
	} else {
		he = &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	code := he.Code
	message := he.Message

	switch m := he.Message.(type) {
	case string:
		message = echo.Map{"message": m}
	case json.Marshaler:
		// do nothing - this type knows how to format itself to JSON
	case error:
		message = echo.Map{"message": m.Error()}
	}

	// Send response
	if c.Request().Method == http.MethodHead { // Issue #608
		err = c.NoContent(he.Code)
	} else {
		err = c.JSON(code, message)
	}
	if err != nil {
		local.Logger.Error("Http Error", zap.Error(err))
	}
}

func Setup(c *local.GContext) *Service<<<ServiceName>>> {

	var (
		pl  *providers.ProvidersLoaded
		err error = nil
	)

	// Normalize serviceHome and make sure it is there
	if i := strings.Index(c.Scfg.ServiceHome, ":"); i > 0 {
		// Mostly because I have to run this on windows too.
		c.Scfg.ServiceHome = c.Scfg.ServiceHome[i+1:]
		c.Scfg.ServiceHome = strings.ReplaceAll(c.Scfg.ServiceHome, "\\", "/")
	}
	if _, err := os.Stat(c.Scfg.ServiceHome); os.IsNotExist(err) {
		panic(fmt.Sprintf("Service home does not exist: %s", c.Scfg.ServiceHome))
	}

	doRegistration(c.Scfg.Name) // This must happen because it might be used by the other setup.
	if err = base.SetupValues(); err != nil {
		panic(fmt.Sprintf("Failed to basic values setup: %v", err))
	}

	// Initial Logger
	glogger = log.NewGLogger(true)
	glogger.AddFileCore(c.Scfg.ServiceHome + "/" + base.LMCONFIG_LOGDIR)
	local.Logger = glogger.GetLogger()
	local.Logger.Info("Logging started.")

	// Additional context setup.
	if err = c.ContextSetup(); err != nil {
		local.Logger.Panic(fmt.Sprintf("Failed to setup context: %v", err))
	}

	// Providers
	if pl, err = providers.ProvidersLoad(c); err != nil {
		local.Logger.Panic(fmt.Sprintf("Failed to load providers: %v", err))
	}

	// Prime echo.
	if echoServer != nil {
		local.Logger.DPanic("Only one server may run per app instance.")
	}
	echoServer = echo.New()

	// Create the service context
	service, errr := NewService<<<ServiceName>>>(c, pl)
	if errr != nil {
		local.Logger.Panic(errr.Error())
	}
	if service == nil {
		local.Logger.Panic("No service created.")
	}

	// Telemetry
	if service.ServiceContext.Telemetry {

		// Get collector.
		var curl string
		curl, err = service.ServiceContext.ConfigProvider.DiscoverService(service.ServiceContext.Scfg.DeploymentName,
			shared.ServiceClassOtelCollector)

		if err == nil {
			if err = gotel.SetupOtel(service.ServiceContext.Scfg.Name, curl); err != nil {
				local.Logger.Error("Failed to set up telemetry", zap.Error(err))
				service.ServiceContext.Telemetry = false
			} else {
				ocfg := gotel.GetOtelConfig()
				if ocfg == nil {
					panic("BUG: Should be impossible for it to be GetOtelConfig here.")
				}
				if ocfg.LoggerProvider != nil {
					//  Add the telemetry sink to logging.
					glogger.AddSlogCore(ocfg.Logger)
					local.Logger = glogger.GetLogger()
				}
			}

		} else {
			service.ServiceContext.Telemetry = false
			if (base.GetErrorFlags(err) & base.ERRFLAG_NOT_FOUND) > 0 {
				local.Logger.Info("Skipping telemetry because no service is running for it.")
			} else {
				local.Logger.Error("Skipping telemetry because of a problem discovering the service class.",
					zap.Error(err))
			}
		}

	} else {
		local.Logger.Info("Telemetry not on.")
	}

	// Additional setup.
	if err = ServiceSetup(service); err != nil {
		local.Logger.Panic(fmt.Sprintf("Failed to setup service: %v", err))
	}

	local.Logger.Info("Service configured.", zap.String(base.LM_SERVICE_NAME, local.ServiceName))
	return service
}

const StRetries = 60
const StInterval = 200

func getServiceInstance(service *Service<<<ServiceName>>>) *config.Config {

	var (
		c    *config.Config
		i    int
		intv = StInterval
	)

	for i = 1; i < StRetries; i++ {
		c = config.GetServiceInstance(service.ServiceContext.ConfigProvider, service.ServiceContext.Scfg.DeploymentName,
			service.ServiceContext.Scfg.Name, service.ServiceContext.Scfg.Id)
		if _, err := c.GetValue(config.CRD_MSERVICE__INSTANCE__IP).GetValueString(); err == nil {
			break
		}
		if i%4 == 0 {
			intv = intv + StInterval
		}
		time.Sleep(time.Duration(StInterval) * time.Millisecond)
	}
	if i >= StRetries {
		c.Err = base.NewGinfraError("Instance is not registered.")
	}

	return c
}

func RunService(service *Service<<<ServiceName>>>) error {
	var (
		err      error = nil
		sw       []*openapi3.T
		sv       []*shared.SchemaValidator
		port     int64
		hostname string
	)
	
	// Logging.
	echoServer.Use(gecho.Echo2ZapLogger(local.Logger))
	echoServer.HTTPErrorHandler = ginfraHTTPErrorHandler

	// Telemetry
	if service.ServiceContext.Telemetry {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			panic("This was binary was not build correctly.  This should never happen.")
		}
		echoServer.Use(gotel.GetEchoMiddleware(bi.Path, "otel-"+service.ServiceContext.Scfg.Name))
		local.Logger.Info("Echo configured to use telemetry.")
	} else {
		local.Logger.Error("Did not set up telemetry for Echo", zap.Error(err))
	}

	// TODO Be careful with this.  We will need to fixup with actual values.
	echoServer.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodHead},
	}))

	// Call registration and make sure the swagger servers are nil.
	sw, err = RegisterHandlers(echoServer, service)
	for _, i := range sw {
		i.Servers = nil
	}

	// Need to handle validation separately.
	if sv, err = shared.GetValidationSchemas(sw); err != nil {
		return err
	}
	echoServer.Use(shared.ValidatorManager(sv))

	// Start service
	local.Logger.Info("Service starting.", zap.String(base.LM_SERVICE_NAME, local.ServiceName))
	scfg := config.GetServiceConfig(service.ServiceContext.ConfigProvider, service.ServiceContext.Scfg.DeploymentName)
	var icfg *config.Config
	if icfg = getServiceInstance(service); icfg.Err != nil {
		return icfg.Err
	}
	if port, err = scfg.GetValue(config.CRD_SSERVICE__CONFIG__PORT).GetValueInt64(); err == nil {
		if hostname, err = icfg.GetValue(config.CRD_MSERVICE__INSTANCE__FQDN).GetValueString(); err == nil {
			if hostname == "127.0.0.1" || hostname == "localhost" {
				hostname = "0.0.0.0"
			}
			err = echoServer.Start(fmt.Sprintf("%v:%v", hostname, port))

			// Wait until the shutdown waitgroup is done, otherwise this will just exit the app.
			shutdownwg.Wait()
		}
	}

	return err
}

func dostop(echoServer *echo.Echo) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := echoServer.Shutdown(ctx)

	if err == nil {
		local.Logger.Info("Shutdown complete.")
	} else {
		local.Logger.Warn("Error occurred during shutdown.", zap.Error(err))
	}

	shutdownwg.Done()
	_ = gotel.ShutdownOtel()
}

func StopService() error {

	if echoServer == nil {
		panic("Service not running.")
	}
	local.Logger.Info("Stopping service.", zap.String(base.LM_SERVICE_NAME, local.ServiceName))
	_ = local.Logger.Sync() // A little late to do anything about it.

	shutdownwg.Add(1)
	go dostop(echoServer)

	echoServer = nil
	return nil // Might pull down the error later.  Not sure it is really needed.
}

// #####################################################################################################################
// # TESTING

const AdminConfigToken = "ABCDEFGHIJ"

type TestServiceContext struct {
	serviceConfig    *config.Config
	named            map[string]string
	testServiceHome  string
	adminConfigToken string
}

func setupTest(t *testing.T, userSetup func(t *testing.T, context *TestServiceContext) error) (
	func(t *testing.T, context *TestServiceContext, teardown func(t *testing.T, context *TestServiceContext) error),
	*TestServiceContext) {

	var (
		tcontext TestServiceContext
		err      error
		port     int64
	)

	// The debugger is being a real pain about this.
	if *testServiceHome == "" {
		sh, _ := os.Getwd()
		testServiceHome = &sh
	}

	// Fake context
	scfg := shared.ServiceConfig{
		ConfigUri:      "file://test/config.json",
		ConfigToken:    AdminConfigToken,
		InstanceType:   config.ITLocal,
		DeploymentName: config.CR_DEPLOYMENT_DEFAULT,
		Name:           local.ServiceName,
		Id:             config.CN_INSTANCE_TYPE__LOCAL,
		Auth:           local.TestAuth,
		ServiceHome:    *testServiceHome,
	}
	gctx := local.GContext{
		Scfg: &scfg,
	}
	local.SetGlobalContext(&gctx)

	service := Setup(&gctx)
	tcontext.serviceConfig = config.GetServiceConfig(service.ServiceContext.ConfigProvider, "default")
	tcontext.testServiceHome = *testServiceHome
	tcontext.adminConfigToken = AdminConfigToken
	if port, err = tcontext.serviceConfig.GetValue(config.CRD_SSERVICE__CONFIG__PORT).GetValueInt64(); err == nil {
		err = userSetup(t, &tcontext)
	}

	if err != nil {
		panic(fmt.Sprintf("SETUP FAILED: %v", err))
	}

	// Run it
	cerr := make(chan error)
	go func() {
		cerr <- RunService(service)
	}()

	// Wait for server.
	hostname, _ := os.Hostname()
	attempts := 0
	for {
		_, err := net.Dial("tcp", fmt.Sprintf(hostname+":%v", port))
		if err == nil {
			break
		}
		attempts++
		if attempts > 10 {
			panic("FAIL: Server did not start.")
		}
		time.Sleep(time.Second)
	}

	// Return a function to teardown the test
	fn := func(t *testing.T, tcontext *TestServiceContext, teardown func(t *testing.T, c *TestServiceContext) error) {

		// Teardown errors are purely informational.
		err := teardown(t, tcontext)
		if err != nil {
			fmt.Printf("ERROR during teardown:  %v", err)
		}

		_ = StopService() // Just consume it for now.
		if errr := <-cerr; errr != nil && !errors.Is(errr, http.ErrServerClosed) {
			panic(fmt.Sprintf("Server closed with error.  %v", errr))
		}
	}

	return fn, &tcontext
}
