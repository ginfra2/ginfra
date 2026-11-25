/*
Package tool
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Support for interacting with a test host that implements service_test interface.
*/
package tool

import (
	"context"
	"fmt"
	"gitlab.com/ginfra/ginfra/api/service_test/client"
	"gitlab.com/ginfra/ginfra/api/service_test/models"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// #####################################################################################################################
// # TEST COMMANDS

func loadConfigProv() (config.ConfigProvider, string, error) {
	var (
		hosted manage.HostFile
		cpv    config.ConfigProvider
	)

	// See if a host file can help us get to a config provider.
	cwd, err := os.Getwd()
	if err == nil {
		if _, err = os.Stat(filepath.Join(cwd, app.PathHostingFilePath)); err == nil {
			// File exists
			if hosted, err = common.ReadJsonFile[manage.HostFile](filepath.Join(cwd, app.PathHostingFilePath),
				"hosting metadata"); err == nil {

				// Get a config provider.
				cpv, err = manage.GetConfigProvider(hosted.ConfigUrl, hosted.ConfigAccessToken)
				return cpv, hosted.DeploymentName, err
			}
		}
	} else {
		err = base.NewGinfraErrorChild("Could not get config provider.", err)
	}

	return cpv, hosted.DeploymentName, err
}

func seekAuthAndHost(gctx *app.GContext) (string, string) {

	var (
		cpv        config.ConfigProvider
		err        error
		deployment string

		auth = gctx.Authtoken
		svc  = gctx.Service
		host = gctx.Host
	)

	if svc == "" {
		svc = app.TestExpectedServiceName
	}

	if auth == "" {
		cpv, deployment, err = loadConfigProv()
		if err == nil {
			cfg := config.GetServiceAuthN(cpv, deployment, svc)
			if err = cfg.GetValueError(); err == nil {
				if cfg.GetValueStringNoerr() != "" {
					auth = cfg.GetValueStringNoerr()
				}
			}
		}

		if err != nil {
			gctx.GetLogger().Warn("Could not get auth token from configuration provider.",
				zap.Error(err))
		}
		if auth == "" {
			return "", ""
		}
	}

	if host == "" {
		if cpv == nil {
			cpv, deployment, err = loadConfigProv()
		}
		if err == nil {
			host, err = cpv.DiscoverPortal(deployment, svc)
		}
		if err != nil {
			gctx.GetLogger().Warn("Could not get host from configuration provider.",
				zap.Error(err))
		}
	}

	return auth, host
}

func TestRun(gctx *app.GContext, uid, target, path string) error {

	var (
		r          *client.GiTestRunResponse
		auth, host = seekAuthAndHost(gctx)
		err        error
	)

	if auth == "" || host == "" {
		return base.NewGinfraError("Authorization and host are required.")
	}

	// Load snippet.
	s, err := os.ReadFile(path)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not load test from file.", err, base.LM_FILEPATH, path)
	}
	ss := string(s)

	var ssp string
	ssp, err = common.SimpleReplacer(ss, map[string]interface{}{"TARGET": target}, true)

	lc, err := client.NewClientWithResponses(host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to host.", err)
	}

	p := models.GiTestRunParams{
		Auth:   auth,
		Target: target,
		Name:   filepath.Base(path),
		Uid:    uid,
	}
	r, err = lc.GiTestRunWithBodyWithResponse(context.Background(), &p, "text/plain",
		strings.NewReader(ssp))

	if err != nil {
		return base.NewGinfraErrorChild("Error connecting to test host.", err)
	}
	if r.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return base.NewGinfraError("Authorization is required.")
	} else if r.HTTPResponse.StatusCode == http.StatusForbidden {
		return base.NewGinfraError("Bad User id (uid).")
	} else if r.HTTPResponse.StatusCode == http.StatusBadRequest {
		if r.JSON400 != nil {
			return base.NewGinfraErrorA("Could not run test due to bad request.", base.LM_CAUSE, *r.JSON400.Message)
		}
		return base.NewGinfraErrorChild("Bad request.", err)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Could not run test.", base.LM_CAUSE, *r.JSON500.Message)
	}

	gctx.GetLogger().Infow("Snippet running.", zap.String(base.LM_ID, *r.JSON200.Id))
	return nil
}

func TestStop(gctx *app.GContext, uid, id string) error {

	var (
		r          *client.GiTestStopResponse
		auth, host = seekAuthAndHost(gctx)
	)

	if auth == "" || host == "" {
		return base.NewGinfraError("Authorization and host are required.")
	}

	lc, err := client.NewClientWithResponses(host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to test host.", err)
	}

	p := models.GiTestStopParams{
		Auth: auth,
		Id:   id,
		Uid:  uid,
	}
	r, err = lc.GiTestStopWithResponse(context.Background(), &p)

	if err != nil {
		return base.NewGinfraErrorChild("Error connecting to test host.", err)
	}
	if r.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return base.NewGinfraError("Authorization is required.")
	} else if r.HTTPResponse.StatusCode == http.StatusForbidden {
		return base.NewGinfraError("User id (uid) is forbidden for this test id.")
	} else if r.HTTPResponse.StatusCode == http.StatusNotFound {
		return base.NewGinfraErrorChildA("Id not found.", err, base.LM_ID, id)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Could not get log.", base.LM_CAUSE, *r.JSON500.Message)
	}

	gctx.GetLogger().Infow("Test stop.", zap.String(base.LM_ID, id),
		zap.String(base.LM_OUTPUT, string(r.Body)))

	return nil
}

func TestStatus(gctx *app.GContext, uid, id string) error {

	var (
		r          *client.GiTestStatusTestResponse
		auth, host = seekAuthAndHost(gctx)
	)

	if auth == "" || host == "" {
		return base.NewGinfraError("Authorization and host are required.")
	}

	lc, err := client.NewClientWithResponses(host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to test host.", err)
	}

	p := models.GiTestStatusTestParams{
		Auth: auth,
		Id:   id,
		Uid:  uid,
	}
	r, err = lc.GiTestStatusTestWithResponse(context.Background(), &p)

	if err != nil {
		return base.NewGinfraErrorChild("Error connecting to test host.", err)
	}
	if r.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return base.NewGinfraError("Authorization is required.")
	} else if r.HTTPResponse.StatusCode == http.StatusForbidden {
		return base.NewGinfraError("User id (uid) is forbidden for this test id.")
	} else if r.HTTPResponse.StatusCode == http.StatusNotFound {
		return base.NewGinfraErrorChildA("Id not found.", err, base.LM_ID, id)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Could not get status.", base.LM_CAUSE, *r.JSON500.Message)
	}

	token := string(*r.JSON200.Token)
	gctx.GetLogger().Infow("Test status", zap.String(base.LM_STATUS, token),
		zap.String(base.LM_CAUSE, *r.JSON200.Status))

	return nil
}

func TestAll(gctx *app.GContext, uid string) error {

	var (
		r          *client.GiTestStatusAllResponse
		auth, host = seekAuthAndHost(gctx)
	)

	if auth == "" || host == "" {
		return base.NewGinfraError("Authorization and host are required.")
	}

	lc, err := client.NewClientWithResponses(host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to test host.", err)
	}

	p := models.GiTestStatusAllParams{
		Auth: auth,
		Uid:  uid,
	}
	r, err = lc.GiTestStatusAllWithResponse(context.Background(), &p)

	if err != nil {
		return base.NewGinfraErrorChild("Error connecting to test host.", err)
	}
	if r.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return base.NewGinfraError("Authorization is required.")
	} else if r.HTTPResponse.StatusCode == http.StatusForbidden {
		return base.NewGinfraError("Bad User id (uid).")
	} else if r.HTTPResponse.StatusCode == http.StatusBadRequest {
		return base.NewGinfraErrorChild("Bad request.", err)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Could not get status.", base.LM_CAUSE, *r.JSON500.Message)
	}

	if gctx.Blob {
		fmt.Println(string(r.Body))
	} else {
		if r.JSON200 != nil {
			if r.JSON200.Running != nil {
				for _, item := range *r.JSON200.Done {
					gctx.GetLogger().Infow("Running test.", zap.String(base.LM_ID, *item.Id),
						zap.String(base.LM_NAME, *item.Name), zap.String(base.LM_STATUS, string(*item.Token)),
						zap.String(base.LM_CAUSE, *item.Status))
				}
			}
			if r.JSON200.Done != nil {
				for _, item := range *r.JSON200.Done {
					gctx.GetLogger().Infow("Done test.", zap.String(base.LM_ID, *item.Id),
						zap.String(base.LM_NAME, *item.Name), zap.String(base.LM_STATUS, string(*item.Token)),
						zap.String(base.LM_CAUSE, *item.Status))
				}
			}
			if r.JSON200.Queue != nil {
				for _, item := range *r.JSON200.Done {
					gctx.GetLogger().Infow("Queued test.", zap.String(base.LM_ID, *item.Id),
						zap.String(base.LM_NAME, *item.Name), zap.String(base.LM_STATUS, string(*item.Token)),
						zap.String(base.LM_CAUSE, *item.Status))
				}
			}
		}
	}

	return nil
}

func TestLog(gctx *app.GContext, uid, id string) error {

	var (
		r          *client.GiTestLogResponse
		auth, host = seekAuthAndHost(gctx)
	)

	if auth == "" || host == "" {
		return base.NewGinfraError("Authorization and host are required.")
	}

	lc, err := client.NewClientWithResponses(host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to test host.", err)
	}

	p := models.GiTestLogParams{
		Auth: auth,
		Id:   id,
		Uid:  uid,
	}
	r, err = lc.GiTestLogWithResponse(context.Background(), &p)

	if err != nil {
		return base.NewGinfraErrorChild("Error connecting to test host.", err)
	}
	if r.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return base.NewGinfraError("Authorization is required.")
	} else if r.HTTPResponse.StatusCode == http.StatusForbidden {
		return base.NewGinfraError("User id (uid) is forbidden for this test id.")
	} else if r.HTTPResponse.StatusCode == http.StatusNotFound {
		return base.NewGinfraErrorChildA("Id not found.", err, base.LM_ID, id)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Could not get log.", base.LM_CAUSE, *r.JSON500.Message)
	}

	if gctx.Blob {
		fmt.Println(string(r.Body))
	} else {
		gctx.GetLogger().Infow("Log file for test.", zap.String(base.LM_ID, id),
			zap.String(base.LM_OUTPUT, string(r.Body)))
	}

	return nil
}

func TestRemove(gctx *app.GContext, uid string, id ...string) error {

	var (
		auth, host = seekAuthAndHost(gctx)
	)

	if auth == "" || host == "" {
		return base.NewGinfraError("Authorization and host are required.")
	}

	if len(id) < 1 {
		return base.NewGinfraError("Id is required.")
	}

	lc, err := client.NewClientWithResponses(host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to test host.", err)
	}

	p := models.GiTestRemoveParams{
		Auth: auth,
		Uid:  uid,
	}

	b := models.TestRemoveRequest{
		Remove: &id,
	}

	r, err := lc.GiTestRemoveWithResponse(context.Background(), &p, b)

	if err != nil {
		return base.NewGinfraErrorChild("Error connecting to test host.", err)
	}
	if r.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return base.NewGinfraError("Authorization is required.")
	} else if r.HTTPResponse.StatusCode == http.StatusForbidden {
		return base.NewGinfraError("Bad User id (uid).")
	} else if r.HTTPResponse.StatusCode == http.StatusNotFound {
		return base.NewGinfraErrorChildA("Id not found.", err)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Could not remove,", base.LM_CAUSE, *r.JSON500.Message)
	}

	gctx.GetLogger().Infow("Remove successful.")

	return nil
}
