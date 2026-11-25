//go:build functional_test

/*
Package manage
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Start local processes tests.
*/
package local

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"path/filepath"
	"testing"
)

func init() {
	common.SetupTest()
}

func testInvalidEportsPanic(gctx *app.GContext) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	var (
		e = make(map[string]shared.Expose)
	)
	e["xxx"] = shared.Expose{
		Name: "Name",
		ServiceClass: shared.ServiceClass{
			Name: "SC Name",
			Url:  "http://<<<host sdfsdfdf",
		},
	}

	cprv, _ := config.GetConfigProvider("file://config.json", "AAAAAAAAAAAA")
	addDiscoveryForExports(gctx, cprv, e)
	return err
}

func TestStartLocalCases(t *testing.T) {
	gctx := app.GetCopyRootContext()
	gctx.ServiceHomePath = filepath.Join(base.GetGinfraHome(), "test", "local", "stub_service")

	t.Run("Bad config provider.", func(t *testing.T) {
		err := StartLocal(&gctx, "jjj://sddsfsd", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Could not load configuration provider")
	})

	gctx.ServiceHomePath = filepath.Join(base.GetGinfraHome(), "test", "local", "service_dirs", "ok_specifics")
	t.Run("Odd settings but ok.", func(t *testing.T) {
		gctx.Clear = true
		gctx.DontStart = true

		// Since it is DontStart it should be ok that the start script isn't there
		err := StartLocal(&gctx, "file://config.json", []string{})
		assert.NoError(t, err)
	})

	// --- Specifics related  -----------------------------------------------------------------------
	gctx.ServiceHomePath = filepath.Join(base.GetGinfraHome(), "test", "local", "service_dirs", "bad_specifics")
	t.Run("Bad command line specifics.", func(t *testing.T) {
		gctx.Specifics = "{\"sdfsdf\":}"
		err := StartLocal(&gctx, "file://config.json", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Specifics (container) could not be loaded from the service")
	})

	t.Run("local container specifics bad.", func(t *testing.T) {
		gctx.Specifics = ""
		err := StartLocal(&gctx, "file://config.json", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Specifics (container) could not be loaded from the service")
	})

	gctx.ServiceHomePath = filepath.Join(base.GetGinfraHome(), "test", "local", "service_dirs", "ok_specifics")
	gctx.Specifics = ""
	t.Run("cannot register", func(t *testing.T) {

		err := StartLocal(&gctx, "error/RegisterService=Could not register/:file://config.json", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Could not register service")
	})

	t.Run("cannot register instance", func(t *testing.T) {
		err := StartLocal(&gctx, "error/RegisterInstance=Could not register/:file://config.json", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Could not register service instance")
	})

	gctx.ServiceHomePath = filepath.Join(base.GetGinfraHome(), "test", "local", "service_dirs", "invalid_specifics")
	t.Run("invalid specifics (expose)", func(t *testing.T) {
		err := testInvalidEportsPanic(&gctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Container specifics are corrupt")
	})
}

func TestStartLocalIndCases(t *testing.T) {
	t.Run("fail to add discovery", func(t *testing.T) {
		var buf bytes.Buffer
		gctx, _ := app.NewContext(&buf)
		cprv, err := config.GetConfigProvider("error/DiscoveryAdd=Failed to add discovery/:memory:/config.json", config.AUTH_ADMIN_TOKEN_DEFAULT)
		if err != nil {
			panic(err)
		}
		addDiscoveryForExports(gctx, cprv, map[string]shared.Expose{"xxxxx": shared.Expose{Name: "yyyy",
			ServiceClass: shared.ServiceClass{Name: "zzzz", Url: "http://localhost"}}})
		assert.Contains(t, buf.String(), "Could not add discovery")
	})

	t.Run("bad service home (panic)", func(t *testing.T) {
		gctx := app.GetCopyRootContext()
		gctx.ServiceHomePath = "/asd292eijwjczxswfadf"
		assert.Panics(t, func() { _ = StartLocal(&gctx, "memory:/", []string{}) })
	})

	t.Run("No meta file", func(t *testing.T) {
		gctx := app.GetCopyRootContext()
		gctx.ServiceHomePath = "/"
		err := StartLocal(&gctx, "memory:/", []string{})
		assert.ErrorContains(t, err, "could not be loaded")
	})

	/*
		Fails on some OSs because root can overwrite the file anyway.
		t.Run("Cannot write service config file.", func(t *testing.T) {
			gctx := app.GetCopyRootContext()
			gctx.ServiceHomePath = filepath.Join(base.GetGinfraTmpTestDir(), "app", "manage", "test", "local")
			err := common.CopyFiles(filepath.Join(base.GetGinfraHome(), "app", "manage", "test", "local"), gctx.ServiceHomePath,
				false, false, false)
			if err != nil {
				panic(err)
			}
			err = os.Chmod(filepath.Join(gctx.ServiceHomePath, base.ConfigFileName), 0444)
			if err != nil {
				panic(err)
			}
			defer func() {
				_ = os.Chmod(filepath.Join(gctx.ServiceHomePath, base.ConfigFileName), 0666)
			}()

			err = StartLocal(&gctx, "config:/", []string{})
			assert.ErrorContains(t, err, "Could not write configuration file")
		})
	*/
}
