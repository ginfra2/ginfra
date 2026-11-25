//go:build unit_test

/*
Package manage
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Tests for common tools and utilities.
*/
package manage

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	ctest "gitlab.com/ginfra/ginfra/common/test"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"os"
	"path/filepath"
	"testing"
)

func TestManageSpecificsGood(t *testing.T) {
	gctx := app.GContext{
		Specifics: `{"spec1": "spec1value", "spec2": "spec2value"}`,
	}

	smap, err := GetServiceSpecifics(&gctx, []string{"spec3", "spec3value"})
	if err != nil {
		t.Error(err)
		return
	}
	ctest.TestCheckMapValue(t, false, smap, "spec1value", "spec1")
	ctest.TestCheckMapValue(t, false, smap, "spec2value", "spec2")
	ctest.TestCheckMapValue(t, false, smap, "spec3value", "spec3")
	ctest.TestCheckMapValue(t, true, smap, "spec3", "spec3value")
}

func TestManageSpecificsBad1(t *testing.T) {
	_ = base.LoadGinfra()

	gctxbad1 := app.GContext{
		Specifics: `{"spec1": {, "spec2": "spec2value"}`,
	}

	_, err := GetServiceSpecifics(&gctxbad1, []string{"spec3", "spec3value"})
	assert.Error(t, err, "expecting broken specifics json")

	gctxbad2 := app.GContext{
		Specifics: `{"spec1": "asdfasdf", "spec2": "spec2value"}`,
	}
	_, err = GetServiceSpecifics(&gctxbad2, []string{"spec3", "spec3value", "unpaired"})
	assert.Error(t, err, "expecting broken specifics additional.")

}

func TestManageCases(t *testing.T) {
	var (
		err error
		cpv config.ConfigProvider
	)

	t.Run("Get config provider bad.", func(t *testing.T) {
		_, err = GetConfigProvider("xxx://yyy", "Q")
		assert.Error(t, err)
	})

	t.Run("Get config provider.", func(t *testing.T) {
		cpv, err = GetConfigProvider("memory://", config.AUTH_ADMIN_TOKEN_DEFAULT)
		if err != nil {
			t.Error(err)
			return
		}
		assert.NotNil(t, cpv)
	})

	t.Run("Rollback registration.", func(t *testing.T) {
		_, err = cpv.RegisterService(config.ITLocal, "test", "test", "test", "test",
			"memory://", "test", 8080, map[string]interface{}{})
		if err != nil {
			t.Error(err)
			return
		}

		var buf bytes.Buffer
		var gctx, _ = app.NewContext(&buf)

		RollbackRegistration(gctx, "test", "test", "test", cpv)
		sp, _ := testutil.NewScanPack(buf.String(), nil)
		assert.Greater(t, sp.Seek("Rolling back"), 1)
		assert.Equal(t, sp.Seek("Failed"), -1)
	})

	t.Run("Rollback registration fail.", func(t *testing.T) {
		var buf bytes.Buffer
		var gctx, _ = app.NewContext(&buf)
		RollbackRegistration(gctx, "xxx", "xxx", "xxx", cpv)
		sp, _ := testutil.NewScanPack(buf.String(), nil)
		assert.Greater(t, sp.Seek("iled to deregister"), 1)
	})

	var tp string
	t.Run("LoadHostFileCwd good file.", func(t *testing.T) {
		tp = base.GetGinfraRandomTmpTestDirAndMake()
		if err = common.CopyFile(filepath.Join(base.GetGinfraHome(), "app", "manage", "test", "host", "good", "_Hosting.json"),
			filepath.Join(tp, "Hosting.json")); err != nil {
			t.Fatal(err)
			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		if err = os.Chdir(tp); err != nil {
			t.Fatal(err)
			return
		}
		defer os.Chdir(cwd)

		_, err = LoadHostFileCwd()
		assert.NoError(t, err)
	})

	t.Run("LoadHostFileCwd bad file.", func(t *testing.T) {
		if err = common.CopyFile(filepath.Join(base.GetGinfraHome(), "app", "manage", "test", "host", "bad", "_Hosting.json"),
			filepath.Join(tp, "Hosting.json")); err != nil {
			t.Fatal(err)
			return
		}
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		if err = os.Chdir(tp); err != nil {
			t.Fatal(err)
			return
		}
		defer os.Chdir(cwd)

		_, err = LoadHostFileCwd()
		assert.ErrorContains(t, err, "Hostfile.json could not be loaded.")
	})
}

func TestPrepSpecifics(t *testing.T) {
	t.Run("Good - default web port", func(t *testing.T) {
		gctx := app.GetRootContext()
		cspec := shared.Specifics{
			Meta: shared.Meta{
				Name: "test",
			},
			Exposed: map[string]shared.Expose{
				"8888": {
					Name: "test",
					ServiceClass: shared.ServiceClass{
						Name: "test",
						Url:  "test",
					},
				},
			},
		}
		ps, err := PrepSpecifics(gctx, &cspec, map[string]interface{}{})
		assert.NoError(t, err)
		assert.NotNil(t, ps)
		assert.Equal(t, "test", ps.Meta.Name)
	})
}
