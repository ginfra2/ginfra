/*
Package host
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Configuration related manage hosted deployments.
*/
package host

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"os"
)

func HostedStore(gctx *app.GContext, name, path, deployment, service string) (err error) {
	_, cprv, err := prepHosted(gctx)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not read source file.", err, base.LM_FILEPATH, path)
	}

	return cprv.StorePut(name, deployment, service, data)
}

func HostedPut(gctx *app.GContext, name, value string) (err error) {
	h, cprv, err := prepHosted(gctx)
	if err != nil {
		return err
	}

	return cprv.PutValue(h.DeploymentName, name, value)
}
