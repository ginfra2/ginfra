/*
Package tool
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

General application information.
*/
package tool

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
)

// #####################################################################################################################
// #

func Info(gctx *app.GContext) error {
	gctx.GetLogger().Infof("Version: %s", base.Version)
	return nil
}
