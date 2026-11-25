/*
Package kube
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes deployments.
*/
package kube

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
)

// #####################################################################################################################
// #

// prepKube returns the populated Hosted object, ConfigProvider object, ConfigProvider URI, and error.
// It reads hosting metadata from a JSON file and fetches the ConfigProvider based on the provided information.
// Input:
// - gctx: The GContext object containing deployment information and logger.
// Output:
// - kube: The Kube object with deployment information.
// - prov: The ConfigProvider object.
// - err: The error, if any.
func prepKube(gctx *app.GContext) (kube manage.HostFile, prov config.ConfigProvider, err error) {
	// -- Get deployment info.
	kube, err = common.ReadJsonFile[manage.HostFile](manage.GetHostFilePath(gctx), "kube hosting metadata")
	if err != nil {
		err = base.NewGinfraErrorChildA(`Could not load kube hosting metadata file (Hosting.json).  Was the deployment 
initialized, is the current working directory not the directory that you initialized it in or is it corrupt?`,
			err, base.LM_FILEPATH, app.PathHostingFilePath)
	} else {
		prov, err = manage.GetConfigProvider(kube.ConfigPortalHostUrl, kube.ConfigAccessToken)
	}
	return kube, prov, err
}
