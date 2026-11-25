/*
Package config
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Minimal config service.  The option to change the ports exist because the tests will not release their listen until
the parent is done.  Huge pain, but here we are.
*/
package config

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"gitlab.com/ginfra/ginfra/api/service_giconfig/models"
	"gitlab.com/ginfra/ginfra/api/service_giconfig/server"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type mconfigServiceEntry struct {
	numConfigRequests int
	configService     *MinimalConfigService
}

var (
	requestLock sync.Mutex
	mcsEntries  map[int]mconfigServiceEntry
)

func init() {
	mcsEntries = make(map[int]mconfigServiceEntry)
}

// RequestConfigService starts or tracks a configuration service on the given port.
// It synchronizes requests using a lock to ensure only one service initialization at a time.
// Returns an error if starting or configuring the service fails.
func RequestConfigService(port int) error {
	requestLock.Lock()
	defer requestLock.Unlock()

	var e mconfigServiceEntry
	if ce, ok := mcsEntries[port]; ok {
		e = ce
	}

	var err error
	e.numConfigRequests = e.numConfigRequests + 1
	if e.numConfigRequests == 1 && (e.configService == nil || e.configService.echoServer == nil) {
		fmt.Printf("Start server: %d", e.numConfigRequests)
		e.configService, err = RunService(port)
	}
	if err == nil {
		// Check to see if the server errored out.  I hate sleeps, but I can't really instrument echo right now.
		time.Sleep(75 * time.Millisecond)
		err = e.configService.Err
	}
	if err != nil {
		err = base.NewGinfraErrorChild("Could not request config service", err)
	}

	mcsEntries[port] = e
	return err
}

// ReleaseConfigService releases a MinimalConfigService associated with the given port or decreases its usage count.
// It stops the service if it is no longer needed and removes its entry from the map.
// Returns an error if the service could not be released.
func ReleaseConfigService(port int) (err error) {
	if ce, ok := mcsEntries[port]; ok {

		if ce.numConfigRequests > 0 {
			if ce.numConfigRequests == 1 {
				err = StopService(ce.configService)
				delete(mcsEntries, port)
			} else {
				ce.numConfigRequests = ce.numConfigRequests - 1
				mcsEntries[port] = ce
			}
		}

		if err != nil {
			err = base.NewGinfraErrorChild("Could not release config service", err)
		}
	}
	return err
}

// ===============================================================================================
// SERVER IMPLEMENTATION

type MinimalConfigService struct {
	csp        MinimalConfigStoreProvider
	echoServer *echo.Echo
	homepath   string
	Err        error
}

type MinimalConfigStoreProvider struct {
	Cp   ConfigProvider
	Lock *sync.Mutex
}

type ConfigStorePutPackage struct {
	path string
	data map[string]interface{}
}

func getConfigStoreProvider(sc *MinimalConfigService) MinimalConfigStoreProvider {
	var (
		provider MinimalConfigStoreProvider
		err      error
		lock     sync.Mutex
	)
	provider.Lock = &lock

	if provider.Cp, err = GetConfigProviderFileSavable(sc.homepath, 10, true); err == nil {
		c := NewConfig(nil, nil)
		c.PutValue(CR_TIME, time.Now())
		if err = provider.Cp.PostConfig(c, CR_LIVE); err == nil {
			c.PutValue(CR_AUTH, AUTH_ADMIN_TOKEN_DEFAULT)
			err = provider.Cp.PostConfig(c, "")
		}
	}

	if err != nil {
		panic(err)
	}
	return provider
}

func RunService(port int) (*MinimalConfigService, error) {

	sc := &MinimalConfigService{}

	err := base.LoadGinfra()
	if err != nil {
		panic(err)
	}
	sc.homepath = base.GetGinfraRandomTmpDir()

	err = os.MkdirAll(sc.homepath, 0766)
	if err != nil {
		return nil, err
	}

	sc.csp = getConfigStoreProvider(sc)

	sc.echoServer = echo.New()
	sc.echoServer.HideBanner = true
	sc.echoServer.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodHead},
	}))

	server.RegisterHandlers(sc.echoServer, sc)

	go func(scc *MinimalConfigService) {
		// I may have to morph sc later.
		scc.Err = scc.echoServer.Start(fmt.Sprintf("%v:%v", "localhost", port))
	}(sc)

	return sc, err
}

func StopService(sc *MinimalConfigService) error {
	if sc.echoServer == nil {
		panic("Service not running.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := sc.echoServer.Shutdown(ctx)

	sc.echoServer = nil
	return err
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

func (gis *MinimalConfigService) GiGetSpecifics(ctx echo.Context) error {
	var r = "{}"
	return ctx.JSON(http.StatusOK, r)
}

func successfulQueryResponse(ctx echo.Context, path string, data map[string]interface{}) error {
	var cqr models.ConfigQueryResponse
	cqr.Qpath = &path
	d, _ := json.Marshal(data)
	b := base64.StdEncoding.EncodeToString(d) // I need a pointer to it.
	cqr.Block = &b
	return ctx.JSON(http.StatusOK, &cqr)
}

func processError(ctx echo.Context, err error, path string) error {
	var cerr *base.GinfraError
	if errors.As(err, &cerr) {
		if (cerr.GetFlags() & base.ERRFLAG_NOT_AUTHORIZED) > 0 {
			return postErrorUnauthorizedResponse(ctx, path, cerr)
		} else if (cerr.GetFlags() & base.ERRFLAG_NOT_FOUND) > 0 {
			return postErrorNotFoundResponse(ctx, cerr, path)
		} else if (cerr.GetFlags() & base.ERRFLAG_SUBSYSTEM_SERIOUS) > 0 {
			return postErrorResponse(ctx, http.StatusInternalServerError, err)
		} else {
			return postErrorResponse(ctx, http.StatusBadRequest, cerr)
		}
	}
	return err
}

// GiConfigQueryConfig. (GET /giconfig/query)
func (sc *MinimalConfigService) GiConfigQueryConfig(ctx echo.Context, params models.GiConfigQueryConfigParams) error {
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	c := sc.csp.Cp.FetchConfig(params.Qpath)

	var err error
	if c.Err == nil {
		err = successfulQueryResponse(ctx, params.Qpath, c.Config)
	} else {
		err = processError(ctx, c.Err, fmt.Sprintf("QUERY: %s", params.Qpath))
	}
	return err
}

func giConfigSetConfigHandleItem(sc *MinimalConfigService, ctx echo.Context, item *models.ConfigSetItem) error {

	var (
		datam map[string]interface{}
		err   error
	)

	if datam, err = common.Base64ToJsonMap(*item.Block); err != nil {
		return postErrorResponse(ctx, http.StatusBadRequest, err)
	}
	if err = sc.csp.Cp.PostConfig(NewConfig(datam, nil),
		strings.Split(*item.Path, ".")...); err != nil {
		err = processError(ctx, err, fmt.Sprintf("QUERY: %s", *item.Path))
	}

	return err
}

// GiConfigSetConfig (POST /giconfig/set)
func (sc *MinimalConfigService) GiConfigSetConfig(ctx echo.Context, params models.GiConfigSetConfigParams) error {
	var (
		req models.ConfigServiceSetRequest
	)
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	err := ctx.Bind(&req)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)

	} else {

		// Ignoring failmode for now.
		if len(*req.Items) < 1 {
			return postErrorResponse(ctx, http.StatusBadRequest, base.NewGinfraError("Nothing in the request."))

		} else {
			for _, item := range *req.Items {
				err = giConfigSetConfigHandleItem(sc, ctx, &item)
				if err != nil {
					return err
				}
			}
		}
	}
	return checkAndPostTerminalOk(ctx, err)
}

// Register a service.
// (POST /giconfig/register)
func (sc *MinimalConfigService) GiConfigRegisterService(ctx echo.Context, params models.GiConfigRegisterServiceParams) error {
	var (
		req   models.ConfigServiceRegistrationRequest
		nauth string
	)
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	err := ctx.Bind(&req)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	var deployment string
	if req.Deployment == nil {
		deployment = CR_DEPLOYMENT_DEFAULT
	} else {
		deployment = *req.Deployment
	}

	its, _ := GetInstanceTypeFromString(*req.Itype)

	port := int(*req.Port)
	if req.Image == nil {
		ni := ""
		req.Image = &ni
	}

	// Service class does not have to be set.
	sclass := ""
	if req.Serviceclass != nil {
		sclass = *req.Serviceclass
	}

	// - MAKE IT  -----------------------------------------------------------------------------------------

	spec, err := common.Base64ToJsonMap(*req.Specifics)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	if nauth, err = sc.csp.Cp.RegisterService(its, *req.Name, deployment, *req.Auth, *req.Image,
		*req.Configprovider, sclass, port, spec); err == nil {

		// At this point if there are no errors, we have succeeded.
		var resp models.ConfigServiceRegistrationResponse
		resp.Confauth = &nauth
		resp.Name = req.Name
		err = ctx.JSON(http.StatusOK, resp)

	} else {
		err = processError(ctx, err, fmt.Sprintf("REGISTER: %s / %s", *req.Name, deployment))
	}

	return err
}

// Register a server instance.
// (POST /giconfig/instance)
func (sc *MinimalConfigService) GiConfigRegisterInstance(ctx echo.Context, params models.GiConfigRegisterInstanceParams) error {
	var (
		req  models.ConfigServiceInstanceRequest
		auth string
	)
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	err := ctx.Bind(&req)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	var deployment string
	if req.Deployment == nil {
		deployment = CR_DEPLOYMENT_DEFAULT
	} else {
		deployment = *req.Deployment
	}

	its, _ := GetInstanceTypeFromString(*req.Itype)

	if auth, err = sc.csp.Cp.RegisterInstance(its, *req.Name, deployment, *req.Sauth, *req.Id, *req.Ip,
		*req.Fqdn); err == nil {

		// At this point if there are no errors, we have succeeded.
		var resp models.ConfigServiceRegistrationResponse
		resp.Confauth = &auth
		resp.Name = req.Name
		err = ctx.JSON(http.StatusOK, resp)

	} else {
		err = processError(ctx, err, fmt.Sprintf("REGISTER: %s / %s", *req.Name, deployment))
	}

	return err
}

// Register or update a portal.
// (POST /giconfig/portal/{operation}/{deployment}/)
func (sc *MinimalConfigService) GiConfigPortal(ctx echo.Context, operation models.GiConfigPortalParamsOperation,
	deployment string, params models.GiConfigPortalParams) error {
	var (
		err error
		req models.ConfigPortalRequest
	)
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	err = ctx.Bind(&req)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	if Operation(operation) == OpPut {
		_, err = sc.csp.Cp.PortalAdd(deployment, *req.Sclass, *req.Portal)
		if err == nil {
			_ = ctx.String(http.StatusOK, "OK")
		} else {
			err = processError(ctx, err, fmt.Sprintf("PORTAL ADD: %s : %s", deployment, *req.Sclass))
		}

	} else if Operation(operation) == OpRemove {
		err = sc.csp.Cp.PortalRemove(deployment, *req.Sclass)
		if err == nil {
			_ = ctx.String(http.StatusOK, "OK")
		} else {
			err = processError(ctx, err, fmt.Sprintf("PORTAL REMOVE: %s : %s", deployment, *req.Sclass))
		}

	} else {
		err = processError(ctx, base.NewGinfraErrorA("Unknown portal operation.", base.LM_TOKEN, operation),
			fmt.Sprintf("PORTAL: %s / %s", deployment, *req.Sclass))
	}

	return err
}

// Deregister a service.
// (POST /giconfig/deregister/{name}/{deployment})
func (sc *MinimalConfigService) GiConfigDeregisterService(ctx echo.Context, deployment string, name string,
	params models.GiConfigDeregisterServiceParams) error {
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()

	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)
	err := sc.csp.Cp.DeregisterService(deployment, name)
	if err != nil {
		return processError(ctx, err, fmt.Sprintf("DEREGISTERSERVICE: %s / %s", name, deployment))
	}
	return ctx.String(http.StatusOK, "OK")
}

// Remove config.
// (GET /giconfig/remove)
func (sc *MinimalConfigService) GiRemoveConfig(ctx echo.Context, params models.GiRemoveConfigParams) error {
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	err := sc.csp.Cp.RemoveConfig(strings.Split(params.Qpath, ".")...)
	if err != nil {
		return processError(ctx, err, fmt.Sprintf("REMOVE: %s", params.Qpath))
	}
	return ctx.String(http.StatusOK, "OK")
}

func (sc *MinimalConfigService) GiConfigDiscoveryAdd(ctx echo.Context, params models.GiConfigDiscoveryAddParams) error {
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	err := sc.csp.Cp.DiscoveryAdd(params.Deployment, params.Sclass, params.Uri)
	if err != nil {
		err = processError(ctx, err, fmt.Sprintf("DISCOVERY ADD: %s / %s", params.Deployment, params.Sclass))

	} else {
		err = ctx.String(http.StatusOK, "OK")
	}
	return err
}

// Discover a service.
// (GET /giconfig/discover/{deployment}/{serviceclass})
func (sc *MinimalConfigService) GiDiscoverService(ctx echo.Context, deployment string, serviceclass string,
	params models.GiDiscoverServiceParams) error {
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	uri, err := sc.csp.Cp.DiscoverService(deployment, serviceclass)
	if err != nil {
		err = processError(ctx, err, fmt.Sprintf("DISCOVER SERVICE: %s / %s", deployment, serviceclass))

	} else {
		var resp models.DiscoverServiceResponse
		resp.Uri = &uri
		err = ctx.JSON(http.StatusOK, resp)
	}
	return err
}

// Discover a portal for a deployment.
// (GET /giconfig/discover/portal/{deployment}/{serviceclass})
func (sc *MinimalConfigService) GiDiscoverPortal(ctx echo.Context, deployment string, serviceclass string,
	params models.GiDiscoverPortalParams) error {
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	p, err := sc.csp.Cp.DiscoverPortal(deployment, serviceclass)
	if err != nil {
		p, err = sc.csp.Cp.DiscoverPortal(deployment, CO_ANY_PORTAL)
	}
	if err != nil {
		return processError(ctx, err, fmt.Sprintf("DISCOVER PORTAL.  Portal not available: %s / %s",
			deployment, serviceclass))
	}

	var resp models.DiscoverPortalResponse
	resp.Uri = &p
	err = ctx.JSON(http.StatusOK, resp)

	return err
}

// Get the current service state.
// (GET /giconfig/state/get/{deployment}/{name}/)
func (sc *MinimalConfigService) GiConfigGetState(ctx echo.Context, deployment string, name string,
	params models.GiConfigGetStateParams) error {
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	i, err := sc.csp.Cp.GetState(deployment, name)
	if err != nil {
		return processError(ctx, err, fmt.Sprintf("INIT STATE.  Service not available: %s / %s",
			deployment, name))
	}

	b := common.JsonMapToBase64(i)
	var r models.ConfigGetStateResponse
	r.Block = &b

	return ctx.JSON(http.StatusOK, r)
}

// Get the current initializaton state.
// (GET /giconfig/state/init/{deployment}/{name}/{id}/)
func (sc *MinimalConfigService) GiConfigGetInit(ctx echo.Context, deployment string, name string, id string,
	params models.GiConfigGetInitParams) error {
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	i, err := sc.csp.Cp.GetInitState(deployment, name, id)
	if err != nil {
		return processError(ctx, err, fmt.Sprintf("INIT STATE.  Service not available: %s / %s / %s",
			deployment, name, id))
	}

	var resp models.ConfigGetInitResponse
	ii := int(i)
	resp.State = &ii
	err = ctx.JSON(http.StatusOK, resp)

	return err
}

// Set State.
// (POST /giconfig/state/post/{deployment}/{name}/{id})
func (sc *MinimalConfigService) GiConfigPostState(ctx echo.Context, deployment string, name string, id string,
	params models.GiConfigPostStateParams) error {
	var (
		err error
		req models.ConfigServiceSetStateRequest
		s   map[string]interface{}
	)
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	err = ctx.Bind(&req)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	s, err = common.Base64ToJsonMap(*req.Block)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	err = sc.csp.Cp.PostState(deployment, name, id, s)
	if err != nil {
		return processError(ctx, err, fmt.Sprintf("INIT STATE.  Service not available: %s / %s / %s",
			deployment, name, ctx.RealIP()))
	}

	var r models.ConfigSetStateResponse
	state := int(InitDone)
	r.State = &state
	return ctx.JSON(http.StatusOK, r)
}

// Store a file.
// (POST /giconfig/store/put)
func (sc *MinimalConfigService) GiConfigStorePut(ctx echo.Context, params models.GiConfigStorePutParams) error {

	rc := ctx.Request().Body
	defer func(rc io.ReadCloser) {
		_ = rc.Close()
	}(rc)

	data, err := io.ReadAll(rc)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	// Lock before acquiring the gcontext.
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()

	var dp, sv string
	if params.Deployment != nil {
		dp = *params.Deployment
	}
	if params.Service != nil {
		sv = *params.Service
	}

	err = sc.csp.Cp.StorePut(params.Name, dp, sv, data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to store file.")
	}

	return ctx.String(http.StatusOK, "OK")

}

// Get a file.
// (POST /giconfig/store/get)
func (sc *MinimalConfigService) GiConfigStoreGet(ctx echo.Context, params models.GiConfigStoreGetParams) error {

	// Lock before acquiring the gcontext.
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()

	var dp, sv string
	if params.Deployment != nil {
		dp = *params.Deployment
	}
	if params.Service != nil {
		sv = *params.Service
	}

	data, err := sc.csp.Cp.StoreGet(params.Name, dp, sv, params.Raw)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get store file.")
	}

	// Return the file as a response
	return ctx.Stream(http.StatusOK, "application/octet-stream", bytes.NewReader(data))
}

// Initialize a data source.
// (POST /giconfig/ginfra/datasource/init)
func (sc *MinimalConfigService) GiControlDatasourceInit(ctx echo.Context, params models.GiControlDatasourceInitParams) error {
	_ = ctx.String(http.StatusOK, "OK")
	return nil
}

// Reset the service/server.
// (POST /giconfig/ginfra/manage/reset)
func (sc *MinimalConfigService) GiControlManageReset(ctx echo.Context, params models.GiControlManageResetParams) error {
	_ = ctx.String(http.StatusOK, "OK")
	return nil
}

// Stop the service/server.
// (POST /giconfig/ginfra/manage/stop)
func (sc *MinimalConfigService) GiControlManageStop(ctx echo.Context, params models.GiControlManageStopParams) error {
	_ = ctx.String(http.StatusOK, "OK")
	return nil
}

// Put a value for the domain.
// (POST /configservice/putvalue)
func (sc *MinimalConfigService) GiConfigPutValue(ctx echo.Context, params models.GiConfigPutValueParams) error {

	// Lock before acquiring the gcontext.
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	var req models.GiConfigPutValueJSONRequestBody
	err := ctx.Bind(&req)
	if err != nil {
		return postErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	err = sc.csp.Cp.PutValue(params.Deployment, params.Name, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get put value.")
	}

	// Return the file as a response
	return ctx.String(http.StatusOK, "OK")
}

// Get a value for the domain.
// (GET /configservice/getvalue)
func (sc *MinimalConfigService) GiConfigGetValue(ctx echo.Context, params models.GiConfigGetValueParams) error {
	// Lock before acquiring the gcontext.
	sc.csp.Lock.Lock()
	defer sc.csp.Lock.Unlock()
	sc.csp.Cp.SetConfigAccessAuth(&params.Confauth)

	svc := ""
	if params.Service != nil {
		svc = *params.Service
	}

	cfg := sc.csp.Cp.GetValue(params.Deployment, svc, params.Name)
	if cfg.GetValueError() != nil {
		return processError(ctx, cfg.GetValueError(), fmt.Sprintf("GET VALUE:  Failed: %s / %s",
			params.Deployment, params.Name))
	}

	var resp models.ConfigGetValueResponse
	ii := cfg.GetValueStringNoerr()
	resp.Block = &ii
	return ctx.JSON(http.StatusOK, resp)
}
