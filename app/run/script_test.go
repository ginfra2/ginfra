//go:build unit_test

/*
Package ginterfaces
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Test runners.
*/
package run

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"path/filepath"
	"testing"
)

func init() {
	common.SetupTest()
}

var testScriptProcessOutputCases = []struct {
	name        string
	context     *ScriptContext
	output      *Output
	out         string
	expectedErr string
	expectedLog string
}{
	{
		name: "Output Delay",
		context: &ScriptContext{
			Configs: map[string]interface{}{},
		},
		output: &Output{
			D: []interface{}{
				Delay(5),
			},
		},
		out:         "",
		expectedErr: "",
		expectedLog: "Delay in milliseconds.",
	},
	{
		name: "Output Source",
		context: &ScriptContext{
			Configs: map[string]interface{}{},
		},
		output: &Output{
			D: []interface{}{
				Source(filepath.Join(base.GetGinfraHome(), "app", "run", "test", "1.yaml")),
			},
		},
		out:         "",
		expectedErr: "",
		expectedLog: "",
	},
	{
		name: "Output Source error",
		context: &ScriptContext{
			Configs: map[string]interface{}{},
		},
		output: &Output{
			D: []interface{}{
				Source("../../.../../../.../../../../../../../../../../../../.."),
			},
		},
		out:         "",
		expectedErr: "OUTPUT SOURCE not a valid path.",
		expectedLog: "",
	},
	{
		name: "Output Scan",
		context: &ScriptContext{
			Configs: map[string]interface{}{},
		},
		output: &Output{
			D: []interface{}{
				Scan("BORK"),
			},
		},
		out:         "asdkfhjsandvalnevvvsdv BORK asdpsaodvnvs",
		expectedErr: "",
		expectedLog: "Scan successful",
	},
	{
		name: "Output Scan Error",
		context: &ScriptContext{
			Configs: map[string]interface{}{},
		},
		output: &Output{
			D: []interface{}{
				Scan("YERP"),
			},
		},
		out:         "asdkfhjsandvalnevvvsdv BORK asdpsaodvnvs",
		expectedErr: "OUTPUT SCAN failed",
		expectedLog: "",
	},
	{
		name: "Output NoScan Error",
		context: &ScriptContext{
			Configs: map[string]interface{}{},
		},
		output: &Output{
			D: []interface{}{
				NoScan("BORK"),
			},
		},
		out:         "asdkfhjsandvalnevvvsdv BORK asdpsaodvnvs",
		expectedErr: "Cmd OUTPUT NOSCAN failed",
		expectedLog: "",
	},
	{
		name: "Output Regex Bad",
		context: &ScriptContext{
			Configs: map[string]interface{}{},
		},
		output: &Output{
			D: []interface{}{
				Extract{Regex: "*"},
				Capture{},
			},
		},
		out:         "",
		expectedErr: "Bad RegEx in EXTRACT",
		expectedLog: "",
	},
}

func TestScriptProcessOutput(t *testing.T) {

	var buf bytes.Buffer
	var gctx, _ = app.NewContext(&buf)

	for _, test := range testScriptProcessOutputCases {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			err := processOutput(gctx, test.context, test.output, test.out)

			// Assert
			if test.expectedErr != "" {
				assert.ErrorContains(t, err, test.expectedErr)
			}
		})
	}

}
