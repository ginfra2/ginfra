//go:build functional_test

/*
Package generate
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Unit tests for generate.  Why is this here?  So we can use the cmd as the entrypoint like would be the case in
real world use.
*/
package generate

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	ctest "gitlab.com/ginfra/ginfra/common/test"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	common.SetupTest()
}

func testGenerateAPIServicesSpec(gctx *app.GContext, target, name, ctype, moduleName string) error {
	if err := ctest.FreshTargetDir(target); err != nil {
		return err
	}
	_ = os.Chdir(target)
	return GenerateService(gctx, name, name, name, ctype, moduleName, filepath.Join(base.GetGinfraHome(), "services",
		"_templates"))
}

func TestGenerateApiService(t *testing.T) {
	gctx, err := app.NewContext(nil)

	cwd, err := os.Getwd()
	if err != nil {
		panic("OS could not get working directory.")
	}
	defer func(dir string) {
		_ = os.Chdir(dir)
	}(cwd)
	tpath := filepath.Join(base.GetGinfraTmpTestDir(), "generate", "gen1")

	t.Run("Generate Server", func(t *testing.T) {
		err := testGenerateAPIServicesSpec(gctx, tpath, "gib2b", string(common.SRVTYPE__APIECHO),
			"ginfra.info/ginfra/api/gservice/service_gib2b")
		if err != nil {
			t.Error(err)
		}

		s, _ := common.GetFileAsString("go.mod")
		assert.GreaterOrEqual(t, strings.Index(s, "module ginfra.info/ginfra/api/gservice/service_gib2b"), 0,
			"mod.go not fixed or present.")
		s, _ = common.GetFileAsString("local/values.go")
		assert.Less(t, strings.Index(s, "const ServiceName = \"<<<ServiceName>>>\""), 0,
			"local/values.go not fixed or present.")
	})

	t.Run("Update Server", func(t *testing.T) {
		// Copy cooked files.
		if err := common.CopyFiles(filepath.Join(base.GetGinfraHome(), "app/generate/test/def1"),
			tpath, true, true, true); err != nil {
			t.Error(err)
			return
		}
		err := os.Chdir(tpath)
		if err != nil {
			t.Error(err)
			return
		}

		err = UpdateService(gctx, filepath.Join(base.GetGinfraHome(), "services", "_templates"))
		if err != nil {
			t.Error(err)
			return
		}

		var sp *testutil.ScanPack
		sp, err = testutil.NewScanPack(common.GetFileAsString("go.mod"))
		assert.Equal(t, nil, err, "Testing error", err)
		assert.GreaterOrEqual(t, sp.Seeks(
			"module ginfra.info/ginfra/api/gservice/service_gib2b",
			"require (",
			"require ("),
			0, "go.mod file not properly processed.")

		sp, err = testutil.NewScanPack(common.GetFileAsString("service/service_impl.go"))
		assert.Equal(t, nil, err, "Testing error", err)
		assert.GreaterOrEqual(t, sp.Seeks(
			"\"os\"",
			"// >D",
			"fmt.Printf(\"More ServiceSetup\")",
			"// >E",
			"fmt.Printf(\"thing\")",
			"// # ADDITIONAL CODE",
			"func MoreStuff() {",
		),
			0, "service/service_impl.go file not properly processed.")

		sp, err = testutil.NewScanPack(common.GetFileAsString("providers/service_providers.go"))
		assert.Equal(t, nil, err, "Testing error", err)
		assert.GreaterOrEqual(t, sp.Seeks(
			"// # ADD POINTERS (without",
			"// >B###############",
			"err *error",
			"// >B#######################",
			"// # ADD POINTERS (without *",
			"l   ProvidersLoaded",
			"// # ADD INSTANTIATING AND ",
			"// >C###############",
			"l.err = nil",
			"// >C##########",
			"// # ADD INSTANTIATING",
			"return &l, err"),
			0, "providers/service_providers.go file not properly processed.")

		sp, err = testutil.NewScanPack(common.GetFileAsString("local/values.go"))
		assert.Equal(t, nil, err, "Testing error", err)
		assert.GreaterOrEqual(t, sp.Seeks(
			"ServiceName = \"gib2b\"",
			"// ^^#########################",
			"// # ADDITIONAL CODE GOES BELOW HERE",
			"func crazyFunction() error {",
			"func anotherCrazyFunction() error",
			"}"),
			0, "local/values.go file not properly processed.")

	})
}

func TestGenerateBadType(t *testing.T) {
	c, _ := app.NewContext(nil)
	err := GenerateService(c, "name", "class", "path", "xxxx", "moduleName",
		filepath.Join(base.GetGinfraHome(), "services", "_templates", "badtype"))
	assert.Error(t, err, "Expecting bad type.")
}
