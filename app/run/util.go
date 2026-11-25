/*
Package run
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Orchestration runner.
*/
package run

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/common"
	"go.uber.org/zap"
	"os"
)

func terminateDeferred(gctx *app.GContext, sctx *ScriptContext) {
	for _, c := range sctx.Detached {
		c.SetKilled(true)
		TerminateCommand(gctx, c.GetCommand())
	}
}

func checkHostFile(gctx *app.GContext, sctx *ScriptContext) {
	if !sctx.HostedSeen {

		_, err := os.Stat(manage.GetHostFilePath(gctx))
		if err == nil {
			// File exists
			hosted, err := common.ReadJsonFile[manage.HostFile](manage.GetHostFilePath(gctx), "hosting metadata")
			if err != nil {
				gctx.GetLogger().Warnw("Could not load hosting meta file.  Values will not be available for string processing.",
					zap.Error(err))
			} else {
				// Values
				sctx.ProcessConfig.Deployment = hosted.DeploymentName

				sctx.ConfigProvider, err = manage.GetConfigProvider(hosted.ConfigUrl, hosted.ConfigAccessToken)
				if err != nil {
					gctx.GetLogger().Warnw("Configuration provider could not be loaded.  It will not be available for string processing.",
						zap.Error(err))
				}
			}

			sctx.HostedSeen = true
		}

	}
}
