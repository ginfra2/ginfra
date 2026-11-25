/*
Package tool
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Minimal service.
*/
package tool

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/ginfra/ginfra/api/service_test/models"
	"gitlab.com/ginfra/ginfra/api/service_test/server"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/service"
	"net/http"
)

// ===============================================================================================
// SERVER IMPLEMENTATION

type ErrorType int

const (
	NONEERROR ErrorType = iota
	AUTHERROR
	INTSRVERROR
	NOTFOUNDERROR
	FORBIDDENERROR
)

type testService struct {
	testServer *service.MinimalServer
	error      ErrorType
}

var minimalTestService testService

func RunService(port int) error {
	var err error
	minimalTestService.testServer, err = service.InitServer(port)
	if err == nil {
		server.RegisterHandlers(minimalTestService.testServer.EchoServer, &minimalTestService)
		err = service.RunServer(minimalTestService.testServer)
	}
	return err
}

func StopService() error {
	return service.StopServer(minimalTestService.testServer)
}

// ForceTestError will cause a one-time error to authorization.
func ForceTestError(t ErrorType) {
	minimalTestService.error = t
}

func (ts *testService) getError() ErrorType {
	r := ts.error
	ts.error = NONEERROR
	return r
}

// =======================================================================================
// = SERVICE IMPLEMENTATIONS

// GiControlDatasourceInit Initialize a data source.
// (POST /ginfra/datasource/init)
func (ts *testService) GiControlDatasourceInit(ctx echo.Context, _ models.GiControlDatasourceInitParams) error {
	var r models.ControlDatasourceInitResponse
	return ctx.JSON(http.StatusOK, r)
}

// GiControlManageReset Reset the service/server.
// (POST /ginfra/manage/reset)
func (ts *testService) GiControlManageReset(ctx echo.Context, _ models.GiControlManageResetParams) error {
	return ctx.JSON(http.StatusOK, "{}")
}

// GiControlManageStop Stop the service/server.
// (POST /ginfra/manage/stop)
func (ts *testService) GiControlManageStop(ctx echo.Context, _ models.GiControlManageStopParams) error {
	return ctx.JSON(http.StatusOK, "{}")
}

func doError(ctx echo.Context, et ErrorType) bool {
	switch et {
	case AUTHERROR:
		_ = ctx.JSON(http.StatusUnauthorized, "ERR")
		return true
	case FORBIDDENERROR:
		_ = ctx.JSON(http.StatusForbidden, "ERR")
		return true
	case NOTFOUNDERROR:
		_ = ctx.JSON(http.StatusNotFound, "ERR")
		return true
	case INTSRVERROR:
		_ = service.PostErrorResponse(ctx, http.StatusInternalServerError, base.NewGinfraError("Internal server error"))
		return true
	case NONEERROR:
		break
	default:
		panic("unhandled default case")
	}
	return false
}

// GiGetSpecifics Get specifics for this service
// (GET /ginfra/specifics)
func (ts *testService) GiGetSpecifics(ctx echo.Context) error {
	if doerr := doError(ctx, ts.getError()); doerr {
		return nil
	}
	return ctx.JSON(http.StatusOK, "{}")
}

// GiTestLog Get log for a test.
// (GET /gitest/log)
func (ts *testService) GiTestLog(ctx echo.Context, _ models.GiTestLogParams) error {
	if doerr := doError(ctx, ts.getError()); doerr {
		return nil
	}
	return ctx.String(http.StatusOK, "LOG")
}

// GiTestRemove Remove a list of completed tests.
// (POST /gitest/remove)
func (ts *testService) GiTestRemove(ctx echo.Context, _ models.GiTestRemoveParams) error {
	if doerr := doError(ctx, ts.getError()); doerr {
		return nil
	}
	return ctx.String(http.StatusOK, "OK")
}

// GiTestRun Run a test.
// (POST /gitest/run)
func (ts *testService) GiTestRun(ctx echo.Context, _ models.GiTestRunParams) error {
	if doerr := doError(ctx, ts.getError()); doerr {
		return nil
	}
	var r models.TestRunInfo
	r.Id = new(string)
	*r.Id = "1"
	return ctx.JSON(http.StatusOK, r)
}

// GiTestStatusAll Get status for all tests.
// (GET /gitest/status/all)
func (ts *testService) GiTestStatusAll(ctx echo.Context, _ models.GiTestStatusAllParams) error {
	if doerr := doError(ctx, ts.getError()); doerr {
		return nil
	}
	var r models.TestStatusInfo
	return ctx.JSON(http.StatusOK, r)
}

// GiTestStatusTest Get status for a test.
// (GET /gitest/status/test)
func (ts *testService) GiTestStatusTest(ctx echo.Context, _ models.GiTestStatusTestParams) error {
	if doerr := doError(ctx, ts.getError()); doerr {
		return nil
	}
	var r models.TestStatusItem
	id := "1"
	r.Id = &id
	running := "running"
	r.Status = &running
	token := models.TestStatusItemToken("running")
	r.Token = &token
	r.Result = &models.TestStatusResult{
		Fail:  new(int),
		Pass:  new(int),
		Total: new(int),
	}
	return ctx.JSON(http.StatusOK, r)
}

// GiTestStop Stop a test.
// (GET /gitest/stop)
func (ts *testService) GiTestStop(ctx echo.Context, _ models.GiTestStopParams) error {
	if doerr := doError(ctx, ts.getError()); doerr {
		return nil
	}
	return ctx.String(http.StatusOK, "OK")
}
