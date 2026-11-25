/*
Package kube
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes deployments.  Utilities for ginfra tasks.
*/
package kube

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"os"
)

// #####################################################################################################################
// # SPECIFICS UTILS

// KubeStore stores a file to the configuration service using the specified name, path, deployment, and service parameters.
// It initializes the configuration provider and reads the file data, then uploads the file to the configuration store.
func KubeStore(gctx *app.GContext, name, path, deployment, service string) (err error) {
	_, cprv, err := prepKube(gctx)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not read source file.", err, base.LM_FILEPATH, path)
	}

	return cprv.StorePut(name, deployment, service, data)
}
