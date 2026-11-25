//go:build functional_test

/*
Package ginterfaces
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Test runners.
*/
package run

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func init() {
	common.SetupTest()

	err := common.CopyFiles(filepath.Join(base.GetGinfraHome(), "app", "run", "test", "glob"),
		filepath.Join(base.GetGinfraTmpTestDir(), "glob"), false, false, false)
	if err != nil {
		panic(err)
	}
}

func TestScript1Good(t *testing.T) {
	var buf bytes.Buffer
	var gctx, _ = app.NewContext(&buf)

	err := Run(gctx, filepath.Join(base.GetGinfraHome(), "app/run/test/1.yaml"))
	assert.NoError(t, err, "test/1.yaml")
	if err != nil {
		return
	}
	sp, _ := testutil.NewScanPack(buf.String(), nil)
	op, _ := testutil.NewScanPack(gctx.MostRecentOutput, nil)
	if runtime.GOOS == "windows" {
		assert.Greater(t, sp.Seek("Windows command is present"), 1, "Success message.")
	} else {
		assert.Greater(t, sp.Seek("Unix command is present"), 1, "Success message.")
	}
	assert.Greater(t, op.Seek("ost management"), 0, "Ginfra output.")
}

func TestScript2Bad(t *testing.T) {
	var buf bytes.Buffer
	var gctx, _ = app.NewContext(&buf)

	err := Run(gctx, filepath.Join(base.GetGinfraHome(), "app/run/test/2.yaml"))
	assert.Error(t, err, "test/2.yaml error")
}

func TestScript3WithConfig(t *testing.T) {
	var buf bytes.Buffer
	var gctx, _ = app.NewContext(&buf)

	err := Run(gctx, filepath.Join(base.GetGinfraHome(), "app/run/test/3.yaml"))
	assert.NoError(t, err, "test/3.yaml")
	if err != nil {
		return
	}
	sp, _ := testutil.NewScanPack(buf.String(), nil)
	op, _ := testutil.NewScanPack(gctx.MostRecentOutput, nil)
	if runtime.GOOS == "windows" {
		assert.Greater(t, sp.Seek("Windows command is present"), 1, "Success message.")
	} else {
		assert.Greater(t, sp.Seek("Unix command is present"), 1, "Success message.")
	}
	assert.Greater(t, op.Seek("ost management"), 0, "Ginfra output.")
}

func TestScriptBadYaml(t *testing.T) {
	gctx, _ := app.NewContext(nil)
	err := Run(gctx, filepath.Join(base.GetGinfraHome(), "app/run/test/bad.yaml"))
	assert.Error(t, err, "Bad yaml should cause error.")
}

var testScriptFiles = []struct {
	name, file string
	errtext    []string
	output     []string
}{
	{
		name:    "CMD",
		file:    "app/run/test/case1cmd.yaml",
		errtext: []string{},
		output:  []string{"Command executed"},
	},
	{
		name:    "GINFRA",
		file:    "app/run/test/case2ginfra.yaml",
		errtext: []string{},
		output:  []string{"Version:"},
	},
	{
		name:    "Config file missing",
		file:    "app/run/test/case3config1.yaml",
		errtext: []string{"Bad Config file specified in Meta.  Could not open."},
		output:  []string{},
	},
	{
		name:    "Config file bad",
		file:    "app/run/test/case4config2.yaml",
		errtext: []string{"Invalid entry", "name value pairs"},
		output:  []string{},
	},
	{
		name:    "Config file bad scan",
		file:    "app/run/test/case4config3.yaml",
		errtext: []string{"token too long"},
		output:  []string{},
	},
	{
		name:    "Meta Terminate",
		file:    "app/run/test/case5meta1.yaml",
		errtext: []string{},
		output:  []string{"Killing command"},
	},
	{
		name:    "Meta Bad WorkingDir",
		file:    "app/run/test/case5meta2.yaml",
		errtext: []string{"Bad WorkingDir set in Meta."},
		output:  []string{},
	},
	{
		name:    "Extract",
		file:    "app/run/test/case6extract.yaml",
		errtext: []string{},
		output:  []string{"ginfra command"},
	},
}

func TestScriptCases(t *testing.T) {
	var buf bytes.Buffer
	var gctx, _ = app.NewContext(&buf)

	for _, tcf := range testScriptFiles {
		t.Run(tcf.name, func(t *testing.T) {
			buf.Reset()
			err := Run(gctx, filepath.Join(base.GetGinfraHome(), tcf.file))

			if err == nil {
				for _, it := range tcf.errtext {
					assert.Less(t, 1, strings.Index(err.Error(), it),
						fmt.Sprintf("Text not found: %s", it))
				}
			}
			sp, _ := testutil.NewScanPack(buf.String(), nil)
			for _, it := range tcf.output {
				assert.Less(t, 1, sp.Seek(it),
					fmt.Sprintf("Text not found: %s", it))
			}
		})
	}
}

func TestScriptProcessPing(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() (*httptest.Server, *Ping)
		wantErr   bool
	}{
		{
			name: "SuccessfulPing",
			setupMock: func() (*httptest.Server, *Ping) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				p := &Ping{
					Url:        ts.URL,
					CodeExpect: http.StatusOK,
					Until:      1000,
					Delay:      1,
				}
				return ts, p
			},
			wantErr: false,
		},
		{
			name: "FailurePingInvalidResponse",
			setupMock: func() (*httptest.Server, *Ping) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}))
				p := &Ping{
					Url:        ts.URL,
					CodeExpect: http.StatusOK,
					Until:      1000,
					Delay:      0,
				}
				return ts, p
			},
			wantErr: true,
		},
		{
			name: "FailurePingTimeout",
			setupMock: func() (*httptest.Server, *Ping) {
				p := &Ping{
					Url:        "http://localhost:60000", // unreachable url
					CodeExpect: http.StatusOK,
					Repeat:     2,
					Until:      1, // short until time to force timeout
					Delay:      0,
				}
				return nil, p
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, p := tt.setupMock()
			if ts != nil {
				defer ts.Close()
			}
			gctx, _ := app.NewContext(nil)
			err := processPing(gctx, p)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDoExecWorkflowCmd(t *testing.T) {
	a, _ := app.NewContext(nil)
	c := Command{}

	tests := []struct {
		name    string
		gctx    *app.GContext
		cmd     string
		failOk  bool
		wantErr bool
	}{
		{
			"Valid command, No error expected",
			a,
			"echo 'Hello world'",
			false,
			false,
		},
		{
			"Valid command with failure tolerated, No error expected",
			a,
			"ls /non-existent-directory",
			true,
			true, // The error still comes out.
		},
		{
			"Valid command with failure not tolerated, Error expected",
			a,
			"ls /non-existent-directory",
			false,
			true,
		},
		{
			"Empty command, Error expected",
			a,
			"",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := doExecWorkflowCmd(tt.gctx, &c, tt.cmd, tt.failOk)

			if (err != nil) != tt.wantErr {
				t.Errorf("doExecWorkflowCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecWorkflowCmd(t *testing.T) {
	cwd, _ := os.Getwd()

	tcTG := "THINGGOOD"

	cases := map[string]struct {
		command    *Command
		wantErr    string
		wantOutput string
		log        string
		nolog      string
	}{
		"Valid Command": {
			command: &Command{
				D: []interface{}{
					Cmd("ls"),
				},
				Platform: PAny,
			},
		},
		"FailOK Command": {
			command: &Command{
				D: []interface{}{
					Cmd("notavalidcommand"),
				},
				FailOk:   true,
				Platform: PAny,
			},
		},
		"Failed Command": {
			command: &Command{
				D: []interface{}{
					Cmd("nonexistingcommand"),
				},
				Platform: PAny,
			},
			wantErr: "executable file not found",
		},
		"WorkingDir Command": {
			command: &Command{
				D: []interface{}{
					WorkingDir{Path: cwd},
				},
				Platform: PAny,
			},
		},
		"Invalid WorkingDir Command": {
			command: &Command{
				D: []interface{}{
					WorkingDir{Path: "/nonexistingpath"},
				},
				Platform: PAny,
			},
			wantErr: "Bad workingdir set in cmd",
		},
		"Remove": {
			command: &Command{
				D: []interface{}{
					Remove{[]string{filepath.Join(base.GetGinfraTmpTestDir(), "glob", "1.yaml")}},
				},
				Platform: PAny,
			},
		},
		"Output fail": {
			command: &Command{
				D: []interface{}{
					Output{D: []interface{}{
						Source(filepath.Join(base.GetGinfraTmpTestDir(), "glob", "2.yaml")),
						Scan("XXXXXXXXXX"),
					}},
				},
				Platform: PAny,
			},
			wantErr: "OUTPUT SCAN failed",
			log:     "Output failed",
		},
		"Ping fail": {
			command: &Command{
				D: []interface{}{
					Ping{Url: "http://localhost:77777/"},
				},
				Platform: PAny,
			},
			wantErr: "PING could not connect",
			log:     "Ping failed",
		},
		"Not successful": {
			command: &Command{
				D: []interface{}{
					Ping{Url: "http://localhost:77777/"},
				},
				C: CommonDirectives{
					Success: &tcTG,
				},

				Platform: PAny,
			},
			wantErr: "PING could not connect",
			nolog:   tcTG,
		},
	}

	var buf bytes.Buffer
	var gctx, _ = app.NewContext(&buf)
	sctx := NewScriptContext("") // Create new Script Context

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			buf.Reset()
			err := execWorkflowCmd(gctx, sctx, tc.command)

			if tc.wantErr != "" && err != nil {
				assert.ErrorContains(t, err, tc.wantErr)
			} else if tc.wantErr == "" && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.wantOutput != "" && strings.Index(gctx.MostRecentOutput, tc.wantOutput) >= 0 {
				t.Errorf("Expected output \"%s\", got \"%s\"", tc.wantOutput, gctx.MostRecentOutput)
			}
			if tc.log != "" {
				assert.Contains(t, buf.String(), tc.log)
			}
			if tc.nolog != "" {
				assert.NotContains(t, buf.String(), tc.nolog)
			}

		})
	}

	t.Run("Remove worked", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(base.GetGinfraTmpTestDir(), "glob", "1.yaml")); !errors.Is(err, os.ErrNotExist) {
			t.Error("Remove did not work.")
		}
	})
}

func TestDoExecWorkflowGinfra(t *testing.T) {
	tcDep := "notdefaul"

	var buf bytes.Buffer
	gctx, _ := app.NewContext(&buf)
	sctx := NewScriptContext("")
	sctx.Script = &Script{Meta: Meta{Deployment: &tcDep}}

	testCases := []struct {
		name    string
		g       *Ginfra
		gcmd    string
		wantErr bool
		flaky   bool
	}{
		{
			name:    "successful execution",
			g:       &Ginfra{Detach: false, D: make([]interface{}, 0)},
			gcmd:    "tool info",
			wantErr: false,
		},
		{
			name:    "execution error",
			g:       &Ginfra{Detach: false, D: make([]interface{}, 0)},
			gcmd:    "nonexistentCommand",
			wantErr: true,
		},
		{
			name:    "successful detached execution",
			g:       &Ginfra{Detach: true, D: make([]interface{}, 0)},
			gcmd:    "tool info",
			wantErr: false,
		},
		// Flaky
		//{
		//	name:    "detached execution error",
		//	g:       &Ginfra{Detach: true, D: make([]interface{}, 0)},
		//	gcmd:    "nonexistentCommand",
		//	wantErr: true,
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// No idea why this is flaky in the pipelines, but it is.
			if !(tc.flaky && testutil.SkipFlakyCheck()) {
				_, _, err := doExecWorkflowGinfraCommand(gctx, sctx, tc.g, tc.gcmd)
				if (err != nil) != tc.wantErr {
					t.Errorf("doExecWorkflowGinfraCommand() error = %v, wantErr %v", err, tc.wantErr)
				}
			}
		})
	}

	sc := Script{Meta: Meta{Terminate: true}}
	defer func() {
		buf.Reset()
		handleDetached(gctx, sctx, &sc, nil)
		assert.Contains(t, "Terminating detached", buf.String())

	}()
}

func TestExecWorkflowGinfra(t *testing.T) {
	tcDep := "notdefaul"

	var buf bytes.Buffer
	gctx, _ := app.NewContext(&buf)
	sctx := NewScriptContext("")
	sctx.Script = &Script{Meta: Meta{Deployment: &tcDep}}

	testCases := []struct {
		name      string
		g         *Ginfra
		wantErr   string
		restoreWd bool
	}{ /*
			{
				name:    "fail command",
				g:       &Ginfra{Detach: false, D: []interface{}{Cmd("blablah")}},
				wantErr: "Cmd failed",
			},
			{
				name:      "good workingdir",
				g:         &Ginfra{Detach: false, D: []interface{}{WorkingDir{base.GetGinfraHome()}}},
				wantErr:   "",
				restoreWd: true,
			},
			{
				name:    "bad workingdir",
				g:       &Ginfra{Detach: false, D: []interface{}{WorkingDir{"/asd93i3jnwo0hvwobbowpwobvobwo"}}},
				wantErr: "Bad workingdir",
				//restoreWd: true,
			},
			{
				name:    "remove",
				g:       &Ginfra{Detach: false, D: []interface{}{Remove{[]string{filepath.Join(base.GetGinfraTmpTestDir(), "glob", "2.yaml")}}}},
				wantErr: "",
			},
			{
				name: "output terminate",
				g: &Ginfra{Detach: true, D: []interface{}{
					Cmd("tool snooze 100000"),
					Output{Terminate: true, D: []interface{}{Delay(5)}},
				}},
				wantErr: "",
			},*/
		{
			name: "output failed",
			g: &Ginfra{Detach: false, D: []interface{}{
				Cmd("tool info"),
				Output{D: []interface{}{Scan("BORK")}},
			}},
			wantErr: "OUTPUT SCAN failed",
		},
		{
			name: "ping fail",
			g: &Ginfra{
				D: []interface{}{
					Ping{Url: "http://localhost:77777/"},
				},
			},
			wantErr: "PING could not connect",
		},
	}

	//

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				err error
				wd  string
			)
			if tc.restoreWd {
				if wd, err = os.Getwd(); err != nil {
					panic(err)
				}
			}

			err = execWorkflowGinfra(gctx, sctx, tc.g)
			if tc.wantErr == "" {
				if err != nil {
					t.Error(err)
				}
			} else {
				assert.ErrorContains(t, err, tc.wantErr)
			}

			if tc.restoreWd {
				if err = os.Chdir(wd); err != nil {
					panic(err)
				}
			}
		})
	}
}

func TestExecOutput(t *testing.T) {
	tcDep := "notdefaul"

	var (
		err error
		wd  string
	)
	gctx, _ := app.NewContext(nil)
	sctx := NewScriptContext("")
	sctx.Script = &Script{Meta: Meta{Deployment: &tcDep}}

	testCases := []struct {
		name    string
		o       *Output
		out     string
		wantErr string
		wd      string
	}{
		{
			name: "until source fail",
			o: &Output{D: []interface{}{
				Until(ScanPollInterval + 15),
				Source("/wiibosbnslabohej3fusdfafd9012j3pnfwonfh9whfw0p9hfepnf"),
			}},
			wantErr: "OUTPUT SOURCE not a valid path",
		},
		{
			name: "until without source",
			o: &Output{D: []interface{}{
				Until(1),
				Scan("xxxxxxxxxxx"),
			}},
			wantErr: "unless SOURCE is set",
		},
		{
			name: "until scan failed",
			o: &Output{D: []interface{}{
				Until(ScanPollInterval + 15),
				Source(filepath.Join("app", "run", "test", "1.yaml")),
				Scan("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
			}},
			wantErr: "UNTIL timeout",
			wd:      base.GetGinfraHome(),
		},
		{
			name: "until noscan failed",
			o: &Output{D: []interface{}{
				Until(ScanPollInterval + 15),
				Source(filepath.Join("app", "run", "test", "1.yaml")),
				NoScan("windows"),
			}},
			wantErr: "UNTIL timeout",
			wd:      base.GetGinfraHome(),
		},
		{
			name: "repeat scan failed",
			o: &Output{Repeat: 2, D: []interface{}{
				Source(filepath.Join("app", "run", "test", "1.yaml")),
				Scan("xxxxxxxx"),
			}},
			wantErr: "OUTPUT SCAN failed",
			wd:      base.GetGinfraHome(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wd != "" {
				// we have to do this junk because there is no telling in which directory the test will run.
				if wd, err = os.Getwd(); err != nil {
					panic(err)
				}
				if err = os.Chdir(tc.wd); err != nil {
					panic(err)
				}
			}

			err := processOutput(gctx, sctx, tc.o, tc.out)
			if tc.wantErr == "" {
				if err != nil {
					t.Error(err)
				}
			} else {
				assert.ErrorContains(t, err, tc.wantErr)
			}

			if tc.wd != "" {
				if err = os.Chdir(wd); err != nil {
					panic(err)
				}
			}
		})
	}
}
