//go:build local_test

/*
Package cmd
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Test the commands.
*/
package cmd

import (
	"context"
	"fmt"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/test"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

const MyPort = 9501

func init() {
	common.SetupTest()
}

func goDropFile() {
	// Wait a second and then drop the flag file.
	time.Sleep(1 * time.Second)
	err := os.WriteFile(test.PathFlagFile, []byte{1}, 0666)
	if err != nil {
		_ = fmt.Sprintf("Could not write flag file: %v", err)
	}
}

func TestStartLocal(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic("OS could not get working directory.")
	}
	defer func(dir string) {
		_ = os.Chdir(dir)
	}(cwd)

	err = os.Chdir(filepath.Join(base.GetGinfraHome(), "test/local/stub_service"))
	if err != nil {
		panic("OS could not change working dir to test/local/stub_service.")
	}

	gctx := app.GetCopyRootContext()
	gctx.ServiceHomePath = ""

	if err = config.RequestConfigService(MyPort); err == nil {

		go goDropFile()

		c := NewLocalStart(&gctx)
		c.SetArgs([]string{fmt.Sprintf("http://localhost:%d/", MyPort)})
		err = c.Execute()

		if err != nil {
			t.Errorf("Error starting local service: %v", err)
		}

		if err = config.ReleaseConfigService(MyPort); err != nil {
			t.Errorf("Could not release config service: %v", err)
		}

	}

}

func startTestServer(port int, wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(fmt.Sprintf("HTTP server error: %v", err))
		}
	}()

	return srv
}

func TestPokeCommand(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	srv := startTestServer(MyPort, &wg)
	defer func() {
		srv.Shutdown(context.Background())
		wg.Wait()
	}()

	time.Sleep(100 * time.Millisecond) // Wait for server to start

	t.Run("poke get", func(t *testing.T) {
		gctx := app.GetCopyRootContext()
		cmd := NewPokeGetCommand(&gctx)
		cmd.SetArgs([]string{fmt.Sprintf("http://localhost:%d/", MyPort), ""})
		if err := cmd.Execute(); err != nil {
			t.Errorf("Error executing get command: %v", err)
		}
	})

	t.Run("poke post", func(t *testing.T) {
		gctx := app.GetCopyRootContext()
		cmd := NewPokePostCommand(&gctx)
		cmd.SetArgs([]string{fmt.Sprintf("http://localhost:%d/", MyPort), ""})
		if err := cmd.Execute(); err != nil {
			t.Errorf("Error executing post command: %v", err)
		}
	})
}
