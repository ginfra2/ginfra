//go:build local_test

/*
Package manage
Copyright (c) 2025 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Start local processes tests.
*/
package local

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/common"
	"net"
	"os"
	"testing"
	"time"
)

func init() {
	common.SetupTest()
}

func TestLocalServer(t *testing.T) {

	s, err := RunService(44244)
	defer func(sc *MinimalLocalService) {
		if s != nil {
			err = StopService(sc)
			if err != nil {
				t.Error("Could not stop service", err)
			}
		}
	}(s)
	gctx := app.GetCopyRootContext()

	timeout := time.After(2 * time.Second)
	tick := time.Tick(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			t.Fatal("Service failed to start within timeout")
		case <-tick:
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", 44244), 100*time.Millisecond)
			if err == nil {
				conn.Close()
				goto PortOpen
			}
		}
	}
PortOpen:

	configPath := "file://app/manage/local/test/"
	if _, err := os.Stat("test/config.json"); err == nil {
		configPath = "file://test/"
	}

	t.Run("Data init", func(t *testing.T) {
		err = LocalInitData(&gctx, configPath+"config.json", "minimal_server", "initName")
		assert.NoError(t, err)
	})

	t.Run("Server error", func(t *testing.T) {
		s.SetErrorState(true)
		err = LocalInitData(&gctx, configPath+"config.json", "minimal_server", "initName")
		assert.Error(t, err)
		s.SetErrorState(false)
	})

	t.Run("Cannot connect", func(t *testing.T) {
		err = LocalInitData(&gctx, configPath+"config_bad.json", "minimal_server", "initName")
		assert.ErrorContains(t, err, "Failed to init service")
	})

}
