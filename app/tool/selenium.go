/*
Package tool
Copyright (c) 2023, 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Support for selenium (https://www.selenium.dev/) execution.  This assumes a hosted selenium is available somewhere and
these commands can communicate with it.  It does not operate selenium or any webdriver directly.
*/
package tool

import (
	"context"
	"gitlab.com/ginfra/ginfra/api/service_selenium/client"
	"gitlab.com/ginfra/ginfra/api/service_selenium/models"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/service/selenium"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
)

// #####################################################################################################################
// # SELENIUM COMMANDS

func SeleniumRun(gctx *app.GContext, host, target, path string) error {

	var (
		r *client.PostGiseleniumRunResponse
	)

	// Load snippet.
	s, err := os.ReadFile(path)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not load snippet from file.", err, base.LM_FILEPATH, path)
	}

	lc, err := client.NewClientWithResponses("http://" + host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to selenium host.", err)
	}

	p := models.PostGiseleniumRunParams{
		Target: target,
	}
	r, err = lc.PostGiseleniumRunWithBodyWithResponse(context.Background(), &p, "text/plain",
		strings.NewReader(string(s)))

	if err != nil {
		return base.NewGinfraErrorChild("Error connecting to selenium host.", err)
	}
	if r.HTTPResponse.StatusCode == http.StatusConflict {
		return base.NewGinfraErrorChild("Maximum number of snippet are already running.", err)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Could not run snippet.", base.LM_CAUSE, *r.JSON500.Message)
	}

	gctx.GetLogger().Infow("Snippet running.", zap.String(base.LM_ID, *r.JSON200.Id))
	return nil
}

func SeleniumStatus(gctx *app.GContext, host string) error {

	var (
		r *client.GetGiseleniumStatusResponse
	)

	lc, err := client.NewClientWithResponses("http://" + host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to selenium host.", err)
	}

	p := models.GetGiseleniumStatusParams{
		Id: "x", // Not used yet.
	}
	r, err = lc.GetGiseleniumStatusWithResponse(context.Background(), &p)

	if err != nil {
		return base.NewGinfraErrorChild("Error contacting host for selenium status.", err)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Error getting selenium status.", base.LM_CAUSE, *r.JSON500.Message)
	}
	if r.JSON200 == nil {
		panic("Empty response.  This should never happen.")
	}

	if *r.JSON200.Status == selenium.CurrentStatusError {
		gctx.GetLogger().Infow("Selenium status", zap.String(base.LM_ID, *r.JSON200.Id),
			zap.String(base.LM_STATUS, *r.JSON200.Status), zap.String(base.LM_ERROR, *r.JSON200.LastError))

	} else if *r.JSON200.Status == selenium.CurrentStatusRunning {
		gctx.GetLogger().Infow("Selenium status", zap.String(base.LM_ID, *r.JSON200.Id),
			zap.String(base.LM_STATUS, *r.JSON200.Status))

	} else {
		gctx.GetLogger().Infow("Selenium status", zap.String(base.LM_STATUS, *r.JSON200.Status))
	}

	return nil
}

func SeleniumLog(gctx *app.GContext, host, id string) error {

	var (
		r *client.GetGiseleniumLogResponse
	)

	lc, err := client.NewClientWithResponses("http://" + host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to selenium host.", err)
	}

	p := models.GetGiseleniumLogParams{
		Id: id,
	}
	r, err = lc.GetGiseleniumLogWithResponse(context.Background(), &p)

	if err != nil {
		return base.NewGinfraErrorChild("Error contacting host for selenium log.", err)
	}
	if r.StatusCode() == http.StatusBadRequest {
		return base.NewGinfraErrorA("Could not determine the id.")
	}
	if r.StatusCode() == http.StatusNotFound {
		return base.NewGinfraErrorA("Log for given id not found.", base.LM_ID, id)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Error getting selenium status.", base.LM_CAUSE, *r.JSON500.Message)
	}
	if r.JSON200 == nil {
		panic("Empty response.  This should never happen.")
	}

	gctx.GetLogger().Infow("Selenium status", zap.String(base.LM_STATUS, *r.JSON200.Status),
		zap.String(base.LM_ID, *r.JSON200.Id), zap.String(base.LM_LOG, *r.JSON200.Log))

	return nil
}

func SeleniumResult(gctx *app.GContext, host, id string) error {

	var (
		r *client.GetGiseleniumResultResponse
	)

	lc, err := client.NewClientWithResponses("http://" + host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to selenium result.", err)
	}

	p := models.GetGiseleniumResultParams{
		Id: id,
	}
	r, err = lc.GetGiseleniumResultWithResponse(context.Background(), &p)

	if err != nil {
		return base.NewGinfraErrorChild("Error contacting host for selenium result.", err)
	}
	if r.StatusCode() == http.StatusBadRequest {
		return base.NewGinfraErrorA("Could not determine the id.")
	}
	if r.StatusCode() == http.StatusNotFound {
		return base.NewGinfraErrorA("Result for given id not found.", base.LM_ID, id)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Error getting selenium result.", base.LM_CAUSE, *r.JSON500.Message)
	}
	if r.JSON200 == nil {
		panic("Empty response.  This should never happen.")
	}

	gctx.GetLogger().Infow("Selenium result", zap.String(base.LM_STATUS, *r.JSON200.Result),
		zap.String(base.LM_ID, id))

	return nil
}

func SeleniumStop(gctx *app.GContext, host, id string) error {

	var (
		r *client.GetGiseleniumStopIdResponse
	)

	lc, err := client.NewClientWithResponses("http://" + host)
	if err != nil {
		return base.NewGinfraErrorChild("Unable to get client to selenium host.", err)
	}

	r, err = lc.GetGiseleniumStopIdWithResponse(context.Background(), id)

	if err != nil {
		return base.NewGinfraErrorChild("Error contacting host for selenium stop.", err)
	}
	if r.JSON500 != nil {
		return base.NewGinfraErrorA("Stop failed.", base.LM_CAUSE, *r.JSON500.Message)
	}

	gctx.GetLogger().Info("Stopped")
	return nil
}
