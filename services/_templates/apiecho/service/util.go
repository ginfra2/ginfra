/*
Package service
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Tools.
*/
package service

import (
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/config"
	"<<<Module>>>/local"
	"go.uber.org/zap"
	"time"
)

// InitService goes through the service initialization process.  It will call initfn if it is the first service to do so.
// It will call loadfn is not the first service.  The idea is the first call can set everything up and the second can
// just load the configuration information.
func InitService(service *Service<<<ServiceName>>> , initfn func(service *Service<<<ServiceName>>> ) error,
	loadfn func(service *Service<<<ServiceName>>> ) error) error {

	const (
		POLL_DELAY = 350
		MAX_POLL   = 15
	)

	var (
		s   config.InitState
		err error
		i   int
	)

	for {
		if s, err = service.ServiceContext.ConfigProvider.GetInitState(service.ServiceContext.Scfg.DeploymentName,
			service.ServiceContext.Scfg.Name, ""); err != nil {
			local.Logger.Error("Could not get server state from configuration service.",
				zap.String(base.LM_CAUSE, err.Error()))

		} else {
			if s == config.InitDone {
				loadfn(service)
				break
			} else if s == config.InitNew {
				initfn(service)
				break
			}

		}

		if i >= MAX_POLL {
			return base.NewGinfraError("Timeout waiting for service initialization by another instance. ")
		}
		time.Sleep(POLL_DELAY * time.Millisecond)
		i++
	}

	return err
}




