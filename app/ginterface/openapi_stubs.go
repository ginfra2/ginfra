/*
Package ginterfaces
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Create client/model/server stubs from an openapi spec yank file.  It will first merge any referenced files from the start
file.  It will look both in GINFRA_HOME/ginterfaces and the current working directory for referenced files.
*/
package ginterfaces

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	cutil "github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
)

func openapiGenerateConfigAndStubs(c *codegen.Configuration, outpath, kind, itxPath string) error {

	swag, err := cutil.LoadSwagger(itxPath)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not load swagger spec.", err, base.LM_FILEPATH, itxPath)
	}

	cout, err := codegen.Generate(swag, *c)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not generate stub.", err, base.LM_TYPE, kind,
			base.LM_FILEPATH, itxPath)
	}

	err = os.WriteFile(outpath, []byte(cout), 0o644)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not write stub.", err, base.LM_TYPE, kind,
			base.LM_FILEPATH, outpath)
	}

	return nil
}

func openapiGetConfigs(module string, use5 bool) (model, client, server *codegen.Configuration, err error) {

	m := codegen.Configuration{
		PackageName: "models",
		Generate: codegen.GenerateOptions{
			Models: true,
		},
	}
	m = m.UpdateDefaults()
	if serr := m.Validate(); serr != nil {
		serr = base.NewGinfraErrorChild("Failed to make models configuration for generation.", serr)
		return
	}
	model = &m

	c := codegen.Configuration{
		PackageName: "client",
		Generate: codegen.GenerateOptions{
			Client:       true,
			EmbeddedSpec: true,
		},
		AdditionalImports: []codegen.AdditionalImport{
			{
				Package: module + "/models",
				Alias:   ".",
			},
		},
	}
	c = c.UpdateDefaults()
	if serr := c.Validate(); serr != nil {
		serr = base.NewGinfraErrorChild("Failed to make client configuration for generation.", serr)
		return
	}
	client = &c

	s := codegen.Configuration{
		PackageName: "server",
		Generate: codegen.GenerateOptions{
			Strict:       true,
			EmbeddedSpec: true,
		},
		AdditionalImports: []codegen.AdditionalImport{
			{
				Package: module + "/models",
				Alias:   ".",
			},
		},
	}

	if use5 {
		s.Generate.Echo5Server = true
	} else {
		s.Generate.EchoServer = true
	}

	s = s.UpdateDefaults()
	if serr := s.Validate(); serr != nil {
		serr = base.NewGinfraErrorChild("Failed to make server configuration for generation.", serr)
		return
	}
	server = &s

	return
}

func openapiGenerateSingle(gctx *app.GContext, use5 bool, serviceeName, name, target, rootModule string, noControl bool, names []string) ([]string, error) {
	var (
		overwrite []string
	)

	// New module
	nmod := rootModule + "/" + app.PathApiDir + "/" + name

	// Destination
	d := path.Join(target, app.PathApiDir, name)

	_ = os.RemoveAll(d)
	err := os.MkdirAll(d, os.ModePerm)
	if err != nil {
		return overwrite, base.NewGinfraErrorChildA("Unable to create directory for api stubs", err, base.LM_FILEPATH, d)
	}

	// New merged interface, scrub names.
	mf := path.Join(d, name+".yaml")
	overwrite, err = OpenApiLoadSaveSpecs(gctx.InterfaceDir, serviceeName, mf, noControl, names)
	if err != nil {
		return overwrite, err
	}

	// Prep
	if err = os.MkdirAll(filepath.Join(d, app.PathModelsDir), os.ModePerm); err == nil {
		if err = os.MkdirAll(filepath.Join(d, app.PathServerDir), os.ModePerm); err == nil {
			err = os.MkdirAll(filepath.Join(d, app.PathClientDir), os.ModePerm)
		}
	}
	model, client, server, err := openapiGetConfigs(nmod, use5)
	if err != nil {
		return overwrite, err
	}

	if err = openapiGenerateConfigAndStubs(model, filepath.Join(d, app.PathModelsDir, name+"_models.gen.go"),
		"models", mf); err == nil {

		if err = openapiGenerateConfigAndStubs(client, filepath.Join(d, app.PathClientDir, name+"_client.gen.go"),
			"client", mf); err == nil {

			err = openapiGenerateConfigAndStubs(server, filepath.Join(d, app.PathServerDir, name+"_server.gen.go"),
				"client", mf)
		}
	}

	return overwrite, err
}

// OpenApiGenerateStubs generates OpenAPI stubs for a specified service and target directory in a Go module structure.
// It verifies the API directory, resolves the module name, and generates necessary files like client, server, and models.
// Returns a list of overwritten files or an error if the process fails.
func OpenApiGenerateStubs(gctx *app.GContext, use5 bool, serviceName, name string, noControl bool, names []string) ([]string, error) {

	var (
		err       error
		overwrite []string
	)

	tf := gctx.Target
	if tf == "" {
		tf, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	rmod, err := app.GetFullModule(tf)
	if err != nil {
		fmt.Printf("Error discovering module: %v\n", err)
		fmt.Printf("Target: %s\n", tf)
		return overwrite, errors.New("Module not discovered.  You must run this from the root directory for a module.")
	}

	err = checkApiDir(tf)
	if err != nil {
		fmt.Printf("Error loading module: %v\n", err)
		return overwrite, err
	}

	overwrite, err = openapiGenerateSingle(gctx, use5, serviceName, name, tf, rmod, noControl, names)
	return overwrite, err
}
