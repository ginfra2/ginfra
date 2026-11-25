//go:build functional_test

/*
Package ginterfaces
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Unit tests for openapi stubs.
*/
package ginterfaces

import (
	"go/ast"
	"path/filepath"
	"testing"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	test "gitlab.com/ginfra/ginfra/common/test"
)

func init() {
	common.SetupTest()
}

func TestGinterfaceOpenapiStubsGood(t *testing.T) {
	td := filepath.Join(base.GetGinfraTmpTestDir(), "openapistub", "good")

	gctx, _ := app.NewContext(nil)
	gctx.TurnDebuggingOn()
	gctx.InterfaceDir = base.GetGinfraGinterfacesDir()
	gctx.Target = td

	app.SetupTestNeedModFile(td)

	_, err := OpenApiGenerateStubs(gctx, false, "gb2b", "goodtest", false, []string{"service_gib2b.yaml"})
	if err != nil {
		t.Error(err)
		return
	}

	oAst, err := base.GetGoAST(filepath.Join(td, "api", "goodtest", "client", "goodtest_client.gen.go"))
	if err != nil {
		t.Error(err)
		return
	}

	// Check the output.  Not really going to do much of these because the schema can change.
	assert.Equal(t, "client", oAst.Name.Name, "Package name incorrect")
	assert.Equal(t, "import", oAst.Decls[0].(*ast.GenDecl).Tok.String(),
		"First import should be an import.")

}

func TestGinterfaceOpenapiStubsCases(t *testing.T) {
	gctx := app.GetCopyRootContext()

	t.Run("Module not discovered", func(t *testing.T) {
		gctx.Target = filepath.Join(base.GetGinfraHome(), "tool")
		_, err := OpenApiGenerateStubs(&gctx, false, "gib2b", "xxx", false, []string{"xxxx"})
		assert.ErrorContains(t, err, "Module not discovered")
	})

	t.Run("Api file instead of directory", func(t *testing.T) {
		gctx.Target = filepath.Join(base.GetGinfraHome(), "app", "ginterface", "test", "stubtest")
		err := checkApiDir(gctx.Target)
		assert.ErrorContains(t, err, "File exists")
	})

	t.Run("Unable to create directory directory", func(t *testing.T) {
		if test.SkipFlakyCheck() {
			return
		}
		gctx.Target = filepath.Join("/dev/null/../../../../../../../../../../../../../../../../../../../../../../../../../../../../..")
		_, err := openapiGenerateSingle(&gctx, false, "gib2b", "xxx", gctx.Target, "xxx",
			false, []string{"xxxx"})
		assert.ErrorContains(t, err, "Unable to create directory")
	})

	t.Run("Bad swagger", func(t *testing.T) {
		err := openapiGenerateConfigAndStubs(nil, "", "",
			filepath.Join(base.GetGinfraHome(), "app", "ginterface", "test", "go.mod"))
		assert.Error(t, err)
	})

	t.Run("Bad generate", func(t *testing.T) {
		var c codegen.Configuration
		err := openapiGenerateConfigAndStubs(&c, "", "",
			filepath.Join(base.GetGinfraHome(), "app", "ginterface", "test", "mergetest", "gi_common_1.yaml"))
		assert.Error(t, err)
	})

	t.Run("cant write stub", func(t *testing.T) {
		c := codegen.Configuration{
			PackageName: "server",
			Generate: codegen.GenerateOptions{
				EchoServer:   true,
				Strict:       true,
				EmbeddedSpec: true,
			},
			AdditionalImports: []codegen.AdditionalImport{
				{
					Package: "mypackage.thing/models",
					Alias:   ".",
				},
			},
		}

		err := openapiGenerateConfigAndStubs(&c, "../../.../../../../.../../../../../../../../../../../../../.../../../../../../../../../.../.../.../...", "",
			filepath.Join(base.GetGinfraHome(), "app", "ginterface", "test", "stubtest", "clean.yaml"))
		assert.ErrorContains(t, err, "Could not write stub")
	})

	t.Run("bad api dir", func(t *testing.T) {
		gctx.Target = filepath.Join(base.GetGinfraHome(), "app", "ginterface", "test", "stubtest")
		_, err := OpenApiGenerateStubs(&gctx, false, "blah", "blah", false, []string{"clean.yaml"})
		assert.ErrorContains(t, err, "File exists where the api")
	})

}
