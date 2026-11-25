//go:build functional_test

/*
Package generate
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Unit tests for generate.  Why is this here?  So we can use the cmd as the entrypoint like would be the case in
real world use.
*/
package cmd

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	common.SetupTest()
}

func testGenerateServerSpec(gctx *app.GContext, target, name, ctype, moduleName string) (string, error) {
	rt := filepath.Join(base.GetGinfraTmpTestDir(), "generate", target)

	if err := os.RemoveAll(rt); err != nil {
		return rt, errors.New("Could not remove previous test files (" + rt + ").  Clean out the tmp/test/ dir.")
	}
	_ = os.MkdirAll(rt, fs.ModePerm)
	_ = os.Chdir(rt)

	g := NewGenerateService()
	// [Service name] [Service class] [Module Name]
	g.SetArgs([]string{name, ctype, moduleName})
	return rt, g.Execute()
}

func TestGenerateServer(t *testing.T) {
	var rt string
	gctx, err := app.NewContext(nil)

	cwd, err := os.Getwd()
	if err != nil {
		panic("OS could not get working directory.")
	}
	defer os.Chdir(cwd)

	t.Run("Generate Server", func(t *testing.T) {
		rt, err = testGenerateServerSpec(gctx, "gen2", "mongodb", string(common.SRVTYPE__SERVER),
			"ginfra.info/ginfra/api/gservice/server_mongodb")
		if err != nil {
			t.Error(err)
			return
		}

		var sp *testutil.ScanPack
		sp, err = testutil.NewScanPack(common.GetFileAsString("specifics.json"))
		assert.NoError(t, err, "GEN2: Could not load specifics.json")
		assert.GreaterOrEqual(t, sp.Seeks(
			"name", "mongodb",
			"type", "server",
			"module_name", "ginfra.info/ginfra/api/gservice/server_mongodb",
			"class", "mongodb"),
			0, "GEN2: specifics file not properly processed.")

		sp, err = testutil.NewScanPack(common.GetFileAsString("build/dockertask.yaml"))
		assert.NoError(t, err, "GEN2: Could not load build/dockertask.yam")
		assert.GreaterOrEqual(t, sp.Seeks(
			"GINFRA_DOCKER_REPO",
			"GINFRA_DOCKER_PATH", "mongodb"), // Mangled in testing because we aren't using the root context
			0, "GEN2: build/dockertask.yaml file not properly processed.")

		sp, err = testutil.NewScanPack(common.GetFileAsString("build/test.yaml"))
		assert.NoError(t, nil, err, "GEN2: Could not load build/test.yaml")
		assert.GreaterOrEqual(t, sp.Seeks(
			"manage host init gservices/giconfig",
			"Configuration provider started",
			"manage host start mongodb mongodb"),
			0, "GEN2: build/test.yaml file not properly processed.")
	})

	// Assume we are still in the same dir as the last
	t.Run("Update Server", func(t *testing.T) {
		err = os.Chdir(rt)
		if err != nil {
			t.Error(err)
			return
		}

		g := NewUpdateService()
		err = g.Execute()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Generate VueVite Server", func(t *testing.T) {
		rt, err = testGenerateServerSpec(gctx, "gen3", "webthing", string(common.SRVTYPE__VUEVITETS),
			"ginfra.info/ginfra/api/gservice/webthing")
		if err != nil {
			t.Error(err)
		}
	})

}
