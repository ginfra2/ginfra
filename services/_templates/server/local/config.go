/*
Package local
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Config for this service.
*/
package local

import (
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"strings"

	"go.uber.org/zap"
)

type GContext struct {

	// Services
	ConfigProvider config.ConfigProvider

	// Service Config
	Scfg *shared.ServiceConfig

	// Orders
	Telemetry bool // Run telemetry
	Clear     bool // Clear data stores and start fresh.

	// Additional Parameters
	Additional []string

	// Private
	it string // For command line
}

var globalContext *GContext

func GetContext() (*GContext, error) {
	var err error

	if globalContext == nil {
		var c GContext
		globalContext = &c
		err = c.GetCmdContext()

		// NOTE: we will always have telemetry on for now.
		globalContext.Telemetry = true
	}

	return globalContext, err
}

// SetGlobalContext is primarily for testing.
func SetGlobalContext(gtcx *GContext) {
	if globalContext != nil {
		panic("Glocal context already set.")
	}
	globalContext = gtcx
}

func (gctx *GContext) ContextSetup() error {

	Logger.Info("Loading configuration provider.", zap.String("uri", gctx.Scfg.ConfigUri))
	cp, err := config.GetConfigProvider(gctx.Scfg.ConfigUri, gctx.Scfg.ConfigToken)
	if err != nil && strings.HasPrefix(gctx.Scfg.ConfigUri, "file:") {
		//"file://test/userfile.json"
		uri := "file://" + gctx.Scfg.ServiceHome + "/" + gctx.Scfg.ConfigUri[7:]
		Logger.Info("Uri did not configure.  Fixing uri and trying again.  :%s", zap.String("uri", uri))
		cp, err = config.GetConfigProvider(uri, gctx.Scfg.ConfigToken)
	}
	gctx.ConfigProvider = cp

	// Test the config provider.
	return err
}

// ^^###################################################################################################################
// # ADDITIONAL CODE GOES BELOW HERE
