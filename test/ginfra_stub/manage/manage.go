/*
Package manage

Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage features for ginfra stub.  These directly map to commands in the cli.
*/

package manage

import (
	"errors"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/test"
	"go.uber.org/zap"
	"os"
	"time"
)

func Local(gctx *app.GContext) error {

	// For now, just remove any flag file in the cwd.  And wait for it to appear again.
	err := os.Remove(test.PathFlagFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		gctx.GetLogger().Warn("Could not remove flag file.", zap.Error(err))
	}

	start := time.Now()
	for {
		time.Sleep(100 * time.Millisecond)
		_, err = os.Stat(test.PathFlagFile)
		if err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}

		d := time.Since(start)
		if d.Seconds() > 180 {
			// Timeout
			break
		}
	}

	gctx.GetLogger().Info("Flag file has appeared.  Quitting.")
	return nil
}
