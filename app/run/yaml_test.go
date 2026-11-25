//go:build unit_test

/*
Package ginterfaces
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Test runners.
*/
package run

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	common.SetupTest()
}

func testGetTestFilePath(name string) string {
	return filepath.Join(base.GetGinfraHome(), "app", "run", "test", name)
}

func TestYaml1(t *testing.T) {

	// Because of the nil bugs.
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	s, err := LoadAndParseYaml(testGetTestFilePath("yaml.yaml"))
	assert.NoError(t, err, "test/yaml.yaml")
	assert.Len(t, s.Workflow, 4, "Four workflow items.")
	if err != nil || len(s.Workflow) != 4 {
		return
	}

	assert.False(t, s.Meta.Terminate, "meta.Terminate")
	assert.Nil(t, s.Meta.WorkingDir, "meta.WorkingDir")
	assert.Equal(t, "thisdeployment", *s.Meta.Deployment, "meta.Deployment")
	assert.Nil(t, s.Meta.Config, "meta.Config")

	assert.True(t, strings.HasSuffix(s.SourceFile, "yaml.yaml"), "sourcefile")

	// Item 0
	assert.Nil(t, s.Workflow[0].Command)
	assert.NotNil(t, s.Workflow[0].Ginfra)

	assert.IsType(t, Cmd("a"), s.Workflow[0].Ginfra.D[0])
	assert.Equal(t, "manage local start config:// --clear", string(s.Workflow[0].Ginfra.D[0].(Cmd)))
}

func testYamlIsError(t *testing.T, file string) error {
	_, err := LoadAndParseYaml(testGetTestFilePath(file))
	assert.Error(t, err)
	return err
}

var testYamlIsErrorFiles = []struct {
	name, file string
	text       []string
}{

	{
		name: "Bad file",
		file: "borkbork",
		text: []string{"Could not read script"},
	},
	{
		name: "Bad yaml",
		file: "yamlbad.yaml",
		text: []string{"Bad format"},
	},
	{
		name: "Bad WORKFLOW is not a list (panic)",
		file: "yamlbadworkflowlist.yaml",
		text: []string{"WORKFLOW should be a list"},
	},
	{
		name: "Bad WORKFLOW SPECIALPANICFGORTESTING (panic)",
		file: "yamlspecialpanic.yaml",
		text: []string{"SPECIALPANICFGORTESTING"},
	},
	{
		name: "Bad top level directive",
		file: "yamlbadtoplevel.yaml",
		text: []string{"Bad top-level"},
	},
	{
		name: "Parse errors",
		file: "yamlerrors.yaml",
		text: []string{"Unknown command PING directive",
			"Top-level META section should be an object"},
	},

	{
		name: "CMD error.",
		file: "yamlcmderrors.yaml",
		text: []string{"Unknown platform name", "platformasdfgh",
			"Empty CMD"},
	},
	{
		name: "GINFRA error.",
		file: "yamlginfraerrors.yaml",
		text: []string{"Unknown GINFRA directive.", "GinfraBork",
			"Bad value for output:delay", "poiuytr",
			"Unknown command in OUTPUT directive", "Thing",
			"Bad value for ping:delay", "xxxx",
			"Bad value for ping:until", "yyyy",
			"Bad value for ping:codeexpect", "zzzz"},
	},
}

func TestYamlCases(t *testing.T) {
	for _, tcf := range testYamlIsErrorFiles {
		t.Run(tcf.name, func(t *testing.T) {
			err := testYamlIsError(t, tcf.file)
			if err != nil {
				for _, it := range tcf.text {
					assert.Contains(t, err.Error(), it)
				}
			}
		})
	}
}

func TestYamlFuncCases(t *testing.T) {
	t.Run("guardString", func(t *testing.T) {
		var errs []error
		_, errs = guardString("guardString", nil, errs)
		assert.ErrorContains(t, errs[0], "Empty")
		_, errs = guardString("guardString", int(1), errs)
		assert.ErrorContains(t, errs[1], "Malformed")
	})
	t.Run("guardTruth", func(t *testing.T) {
		var errs []error
		_, errs = guardTruth("guardTruth", nil, errs)
		assert.ErrorContains(t, errs[0], "Empty")
		_, errs = guardTruth("guardTruth", int(1), errs)
		assert.ErrorContains(t, errs[1], "Malformed")
	})
	t.Run("guardInt", func(t *testing.T) {
		var (
			i    int
			errs []error
		)
		i, errs = guardInt("guardInt", int(1), errs)
		assert.Equal(t, 1, i)
		_, errs = guardInt("guardInt", nil, errs)
		assert.ErrorContains(t, errs[0], "Empty")
		_, errs = guardInt("guardInt", true, errs)
		assert.ErrorContains(t, errs[1], "Malformed")
	})
	t.Run("Bad Tokens", func(t *testing.T) {
		var errs []error
		errs = parseYamlMeta(nil, map[string]interface{}{"XXXX": "YYYY"}, errs)
		assert.ErrorContains(t, errs[0], "Bad META section.  Unknown token")
		_, errs = parseYamlExtract(map[string]interface{}{"XXXX": "YYYY"}, errs)
		assert.ErrorContains(t, errs[1], "Unknown command in EXTRACT directive")
		_, errs = parseYamlCapture(map[string]interface{}{TOKEN_TARGET: "xxx", "XXXX": "YYYY"}, errs)
		assert.ErrorContains(t, errs[2], "Unknown command in CAPTURE directive")
	})
	t.Run("Bad Types", func(t *testing.T) {
		var errs []error
		_, errs = parseYamlPing(11, errs)
		assert.ErrorContains(t, errs[0], "PING must be a yaml object")
		_, errs = parseYamlOutput(11, errs)
		assert.ErrorContains(t, errs[1], "Unknown type for OUTPUT")
	})
	t.Run("Token funcs", func(t *testing.T) {
		var errs []error
		_, errs = tokenRemove(nil, errs)
		assert.ErrorContains(t, errs[0], "Empty")
		_, errs = tokenRemove(1, errs)
		assert.ErrorContains(t, errs[1], "Malformed")

		_, errs = tokenPing(nil, errs, true, true)
		assert.ErrorContains(t, errs[2], "Empty")
		_, errs = tokenPing(1, errs, true, false)
		assert.ErrorContains(t, errs[3], "PING before CMD is not allowed")

		_, errs = tokenOutput(nil, errs, true, true)
		assert.ErrorContains(t, errs[4], "Empty")
		_, errs = tokenOutput(1, errs, true, false)
		assert.ErrorContains(t, errs[5], "OUTPUT before CMD is not allowed")
	})
	t.Run("Snippets", func(t *testing.T) {
		var err error

		_, err = LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Cmd:\n      Platform:\n      Cmd: ginfra -v"), "/test")
		assert.ErrorContains(t, err, "Empty PLATFORM")
		_, err = LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Cmd:\n      Platform: 1\n      Cmd: ginfra -v"), "/test")
		assert.ErrorContains(t, err, "Malformed")

		_, err = LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Cmd:\n      Platform:\n      Cmd: \n          This: Thing"), "/test")
		assert.ErrorContains(t, err, "Malformed CMD")

		_, err = LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Cmd: \n      - Morestuff"), "/test")
		assert.ErrorContains(t, err, "Poorly formed yaml CMD")
		_, err = LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Ginfra: \n      - Morestuff"), "/test")
		assert.ErrorContains(t, err, "Poorly formed yaml GINFRA")

		_, err = LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Cmd:\n"), "/test")
		assert.ErrorContains(t, err, "Poorly formed yaml for CMD.")
		_, err = LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Ginfra:\n"), "/test")
		assert.ErrorContains(t, err, "Poorly formed yaml for GINFRA.")

		_, err = LoadAndParseYamlData([]byte("---\nWorkflow: \n  - Ginfra\n"), "/test")
		assert.ErrorContains(t, err, "Expecting a directive in the WORKFLOW list")

		_, err = LoadAndParseYamlData([]byte("---\nWorkflow: \n  - XXX: yyy\n"), "/test")
		assert.ErrorContains(t, err, "Unknown type for XXX")

	})

	t.Run("Odds", func(t *testing.T) {
		var errs []error
		_, errs = parseYamlExtract(map[string]interface{}{"COMMENT": "YYYY"}, errs)
		assert.Equal(t, 0, len(errs), "Comment should be allowed in Extract.")
		_, errs = parseYamlCapture(map[string]interface{}{"COMMENT": "YYYY"}, errs)
		assert.Equal(t, 0, len(errs), "Comment should be allowed in CAPTURE.")

		_, err := LoadAndParseYamlData([]byte("---\nWorkflow:  - Comment: sdf\n\n"), "/test")
		assert.NoError(t, err)
	})

	t.Run("Output object", func(t *testing.T) {
		y, err := LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Ginfra:\n      - Output:\n          Delay: 1\n          NoScan: Bork"), "/test")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(y.Workflow[0].Ginfra.D))
		o := y.Workflow[0].Ginfra.D[0].(Output)
		assert.Equal(t, 2, len(o.D))
	})

	t.Run("Output list", func(t *testing.T) {
		y, err := LoadAndParseYamlData([]byte("---\nWorkflow:\n  - Ginfra:\n      - Output:\n          - Delay: 1\n          - NoScan: Bork"), "/test")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(y.Workflow[0].Ginfra.D))
		o := y.Workflow[0].Ginfra.D[0].(Output)
		assert.Equal(t, 2, len(o.D))
	})

	t.Run("Regression 1", func(t *testing.T) {
		_, err := LoadAndParseYaml(testGetTestFilePath("regression_1.yaml"))
		assert.NoError(t, err)
	})
}
