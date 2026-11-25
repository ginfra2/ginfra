//go:build unit_test

/*
Package cmd
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Test the commands.
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	testutil "gitlab.com/ginfra/ginfra/common/test"
)

func init() {
	common.SetupTest()
}

func TestRoot(t *testing.T) {
	c := NewRootCommand(nil)

	DebuggingFlag = c.PersistentFlags().Bool("debugging", false, "turn on debugging log")

	c.SetArgs([]string{"--debugging"})
	err := c.Execute()

	assert.NoError(t, err)
}

func TestCmdRun(t *testing.T) {
	var buf bytes.Buffer
	var gctx, _ = app.NewContext(&buf)

	c := NewRunCommand(gctx)
	c.SetArgs([]string{filepath.Join(base.GetGinfraHome(), "app/run/test/1.yaml")})
	err := c.Execute()
	if err != nil {
		t.Errorf("Error executing run command: %v", err)
		return
	}

	sp, _ := testutil.NewScanPack(buf.String(), nil)
	assert.Greater(t, sp.Seek("Starting command"), 1)
	assert.Greater(t, sp.Seek("ginfra version"), 1)
	assert.Greater(t, sp.Seek("Scan successful"), 1)
	assert.Greater(t, sp.Seek("ginfra host"), 1)
}

func TestCmdRunPost(t *testing.T) {
	var buf bytes.Buffer
	var gctx, _ = app.NewContext(&buf)

	c := NewRunCommand(gctx)
	c.SetArgs([]string{filepath.Join(base.GetGinfraHome(), "app/cmd/test/1.yaml")})
	err := c.Execute()
	if err != nil {
		t.Errorf("Error executing run command: %v", err)
		return
	}

	sp, _ := testutil.NewScanPack(buf.String(), nil)
	assert.Greater(t, sp.Seek("Starting command"), 1)
	assert.Greater(t, sp.Seek("SCAN successful"), 1)
	assert.Equal(t, -1, sp.Seek("unknown flag: --bork"), "Post was not on.")
	assert.Greater(t, sp.Seek("unknown flag: --duck"), 1)
}

func TestCmdManage(t *testing.T) {
	gctx := app.GetCopyRootContext()
	gctx.ServiceHomePath = filepath.Join(base.GetGinfraHome(), "test", "local", "stub_service")

	c := NewLocalStart(&gctx)
	c.SilenceErrors = true
	c.SetArgs([]string{".", "extra", "junk"}) // This will fail, but that is ok.
	err := c.Execute()

	if err == nil {
		t.Errorf("Expected error executing run command: %v", err)
		return
	}

	sp, _ := testutil.NewScanPack(err.Error(), nil)
	assert.Greater(t, sp.Seek("ssage"), 1)
	fmt.Printf("%v\n", err.Error())
	assert.Greater(t, sp.Seek("Unknown ConfigProvider type"), 1)
}

func TestCmdErrorNoError(t *testing.T) {
	gctx := app.GetCopyRootContext()
	gctx.ServiceHomePath = filepath.Join(base.GetGinfraHome(), "test", "local", "stub_service")

	tctx := app.GetCopyRootContext()
	tctx.Target = "xxxx"

	cwd, err := os.Getwd()
	if err != nil {
		panic("OS could not get working directory.")
	}
	defer func(dir string) {
		_ = os.Chdir(dir)
	}(cwd)
	err = os.Chdir(gctx.ServiceHomePath)
	if err != nil {
		panic(err)
	}

	testCases := []struct {
		name    string
		cmd     *cobra.Command
		args    []string
		errText string
	}{
		{
			name:    "Root command -- doesn't do anything",
			cmd:     NewRootCommand(&gctx),
			args:    []string{},
			errText: "",
		},
		{
			name:    "Generate list ok",
			cmd:     NewGenerateListCmd(),
			args:    []string{},
			errText: "",
		},
		/*
			{
				name:    "Ginterface merge openapi, bad service",
				cmd:     NewGinterfaceMergeCmd(&gctx),
				args:    []string{"openapi", "xxxxx", "xxxx"},
				errText: "Does not exist",
			},
			{
				name:    "Ginterface merge openapi, bad target",
				cmd:     NewGinterfaceMergeCmd(&tctx),
				args:    []string{"openapi", "xxxxx", "xxxx"},
				errText: "Does not exist",
			}, */
		{
			name:    "Tool snooze ok",
			cmd:     NewSnoozeCommand(),
			args:    []string{"10"},
			errText: "",
		},
		{
			name:    "Tool snooze bad value",
			cmd:     NewSnoozeCommand(),
			args:    []string{"sdf82jns9hg#$%"},
			errText: "Bad snooze value",
		},
		{
			name:    "Selenium run",
			cmd:     NewSeleniumRunCommand(),
			args:    []string{"afasdf", "asdfasdf", "asdfasdfa", "asdfasdf"},
			errText: "Could not load snippet",
		},
		{
			name:    "Selenium status",
			cmd:     NewSeleniumStatusCommand(),
			args:    []string{"afasdf"},
			errText: "Error contacting host",
		},
		{
			name:    "Selenium log",
			cmd:     NewSeleniumLogCommand(),
			args:    []string{"afasdf", "asdf"},
			errText: "Error contacting host",
		},
		{
			name:    "Selenium result",
			cmd:     NewSeleniumResultCommand(),
			args:    []string{"afasdf", "asdf"},
			errText: "Error contacting host",
		},
		{
			name:    "Selenium stop",
			cmd:     NewSeleniumStopCommand(),
			args:    []string{"afasdf", "asdfads"},
			errText: "Error contacting host",
		},
		{
			name:    "Ginterface merge",
			cmd:     NewGinterfaceMergeCmd(&gctx),
			args:    []string{"afasdf", "asdasd", "asdfasdf"},
			errText: "Unknown",
		},
		{
			name:    "Ginterface stubs",
			cmd:     NewGinterfaceStubsCmd(&gctx),
			args:    []string{"afas1239ehqihdf", "asdasd", "sdfasdf", "asdfasdf"},
			errText: "Unknown stub type",
		},
		{
			name:    "Ginterface stubs -- openapi not found",
			cmd:     NewGinterfaceStubsCmd(&gctx),
			args:    []string{"openapi", "asdasd", "sdfasdf", "asdfasdf"},
			errText: "Does not exist",
		},
		{
			name:    "Root",
			cmd:     NewRootCommand(nil),
			args:    []string{},
			errText: "",
		},
		{
			name:    "Manage Host Init not enough args",
			cmd:     NewManageHostInitCmd(&gctx),
			args:    []string{},
			errText: "at least 1 arg",
		},
		{
			name:    "Manage Host Start",
			cmd:     NewManageHostStartCmd(&gctx),
			args:    []string{"aaa", "bbb"},
			errText: "Hosting.json",
		},
		{
			name:    "Manage Host List",
			cmd:     NewManageHostListCmd(&gctx),
			args:    []string{},
			errText: "Hosting.json",
		},
		{
			name:    "Manage Host Show",
			cmd:     NewManageHostShowCmd(&gctx),
			args:    []string{"aaaa"},
			errText: "Hosting.json",
		},
		{
			name:    "Manage Host Stop",
			cmd:     NewManageHostStopCmd(&gctx),
			args:    []string{"aaaa"},
			errText: "Hosting.json",
		},
		{
			name:    "Manage Host Init Data - bad service",
			cmd:     NewManageHostInitData(&gctx),
			args:    []string{"xxxx", "xxxxx"},
			errText: "not load",
		},
		{
			name:    "Manage Host Init Data - not enough args",
			cmd:     NewManageHostInitData(&gctx),
			args:    []string{"xxxx"},
			errText: "at least 2 arg",
		},
		{
			name:    "Manage Host Reset",
			cmd:     NewManageHostResetCmd(&gctx),
			args:    []string{"aaaa"},
			errText: "Hosting.json",
		},
		{
			name:    "Manage Host Destroy",
			cmd:     NewManageHostDestroyCmd(&gctx),
			args:    []string{"aaaa"},
			errText: "Hosting.json",
		},
		{
			name:    "Manage Host Store",
			cmd:     NewManageHostStoreCmd(&gctx),
			args:    []string{"aaaa", "bbbb", "cccc", "dddd"},
			errText: "Hosting.json",
		},
		{
			name:    "Manage Host Put",
			cmd:     NewManageHostPutCmd(&gctx),
			args:    []string{"aaaa", "bbbb"},
			errText: "Hosting.json",
		},
		{
			name:    "Manage Local InitData error - bad config provider",
			cmd:     NewLocalInitData(&gctx),
			args:    []string{"aaaa", "bbbb", "xxxx"},
			errText: "Could not load configuration provider",
		},
		{
			name:    "Manage Local InitData error - not registered",
			cmd:     NewLocalInitData(&gctx),
			args:    []string{"file://" + filepath.Join(base.GetGinfraHome(), "common", "config", "test", "1.json"), "bbbb", "xxxx"},
			errText: "Service and deployment not registered.",
		},
		{
			name:    "Manage Local InitData error - no auth",
			cmd:     NewLocalInitData(&gctx),
			args:    []string{"file://" + filepath.Join(base.GetGinfraHome(), "common", "config", "test", "1_noauth.json"), "server_otelcol", "xxxx"},
			errText: "", // Just a log message so no error
		},
		// Not a reliable test since sometimes random things are listening on the port.
		//{
		//	name:    "Manage Local InitData error - not listening",
		//	cmd:     NewLocalInitData(&gctx),
		//	args:    []string{"file://" + filepath.Join(base.GetGinfraHome(), "common", "config", "test", "1.json"), "server_otelcol", "xxxx"},
		//	errText: "connection refused",
		//},
		{
			name:    "Poke post error - not enough args",
			cmd:     NewPokePostCommand(&gctx),
			args:    []string{"xxx"},
			errText: "requires at least 2 arg",
		},
		{
			name:    "Poke post error - uneven args",
			cmd:     NewPokePostCommand(&gctx),
			args:    []string{"xxx", "xxx", "xxx", "xxx", "xxx"},
			errText: "Odd number of name/value arguments",
		},
		{
			name:    "Poke post error - even args",
			cmd:     NewPokePostCommand(&gctx),
			args:    []string{"xxx", "xxx", "xxx", "xxx"},
			errText: "unsupported protocol scheme",
		},
		{
			name:    "Poke post error - no args",
			cmd:     NewPokePostCommand(&gctx),
			args:    []string{"xxx", "xxx"},
			errText: "unsupported protocol scheme",
		},
		{
			name:    "Get get error - not enough args",
			cmd:     NewPokeGetCommand(&gctx),
			args:    []string{"xxx"},
			errText: "requires at least 2 arg",
		},
		{
			name:    "Poke get error - uneven args",
			cmd:     NewPokeGetCommand(&gctx),
			args:    []string{"xxx", "xxx", "xxx", "xxx", "xxx"},
			errText: "Odd number of name/value arguments",
		},
		{
			name:    "Poke get error - even args",
			cmd:     NewPokeGetCommand(&gctx),
			args:    []string{"xxx", "xxx", "xxx", "xxx"},
			errText: "unsupported protocol scheme",
		},
		{
			name:    "Poke get error - no args",
			cmd:     NewPokeGetCommand(&gctx),
			args:    []string{"xxx", "xxx"},
			errText: "unsupported protocol scheme",
		},
		{
			name:    "set specifics - no args, no file",
			cmd:     NewConfigSetSpecificsCommand(&gctx),
			args:    []string{"xxx", "xxx", "xxx"},
			errText: "Hosting.json file not found.",
		},
		{
			name:    "tool test run",
			cmd:     NewTestRunCommand(),
			args:    []string{"xxx", "xxx", "xxx"},
			errText: "Authorization and host are required",
		},
		{
			name:    "tool test stop",
			cmd:     NewTestStopCommand(),
			args:    []string{"xxx", "xxx"},
			errText: "Authorization and host are required",
		},
		{
			name:    "tool test status",
			cmd:     NewTestStatusCommand(),
			args:    []string{"xxx", "xxx"},
			errText: "Authorization and host are required",
		},
		{
			name:    "tool test status all",
			cmd:     NewTestAllCommand(),
			args:    []string{"xxx"},
			errText: "Authorization and host are required",
		},
		{
			name:    "tool test log",
			cmd:     NewTestLogCommand(),
			args:    []string{"xxx", "xxx"},
			errText: "Authorization and host are required",
		},
		{
			name:    "tool test remove",
			cmd:     NewTestRemoveCommand(),
			args:    []string{"xxx", "xxx"},
			errText: "Authorization and host are required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cmd.SetArgs(tc.args) // This will fail, but that is ok.
			err := tc.cmd.Execute()
			if tc.errText == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errText)
			}
		})
	}
}

func TestCmdHostErrorNoError(t *testing.T) {
	gctx := app.GetCopyRootContext()
	gctx.ServiceHomePath = base.GetGinfraTmpTestDir()

	cwd, err := os.Getwd()
	if err != nil {
		panic("OS could not get working directory.")
	}
	defer func(dir string) {
		_ = os.Chdir(dir)
	}(cwd)
	err = os.Chdir(gctx.ServiceHomePath)
	if err != nil {
		panic(err)
	}

	testCases := []struct {
		name   string
		cmd    *cobra.Command
		args   []string
		hasErr bool
	}{
		{
			name:   "Manage Host Init",
			cmd:    NewManageHostInitCmd(&gctx),
			args:   []string{"xxxxx"},
			hasErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cmd.SetArgs(tc.args) // This will fail, but that is ok.
			err := tc.cmd.Execute()
			if tc.hasErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestCmdSetSpecificsWithFile(t *testing.T) {
	if !testutil.SkipFlakyCheck() {
		gctx := app.UseRootContext()
		cwd, err := os.Getwd()
		if err != nil {
			panic("OS could not get working directory.")
		}
		defer func(dir string) {
			app.ReleaseRootContext()
			_ = os.Chdir(dir)
		}(cwd)

		hosting := `{
  "DeploymentName": "default",
  "ConfigAccessToken": "AVIQWLFVBKJE",
  "ConfigUrl": "http://172.18.0.2:5111",
  "ConfigPortalHostUrl": "http://deskgoat:5111",
  "NextServicePort": 8904,
  "ConfigContainerId": "fabe968e52de",
  "DockerHost": "deskgoat",
  "NetworkName": "ginfranet_default",
  "NetworkId": "bb5aae5f66e8",
  "InitConfigMapName": ""
}
`

		gctx.ServiceHomePath = filepath.Join(base.GetGinfraHome(), "app", "cmd", "test")
		err = os.Chdir(gctx.ServiceHomePath)
		if err != nil {
			panic(err)
		}

		if err = os.WriteFile(filepath.Join(gctx.ServiceHomePath, "Hosting.json"), []byte(hosting), 0644); err != nil {
			t.Errorf("Error writing file: %v", err)
			return
		}

		// Completely random behavior across six separate servers. This command shit is just absolute trash.
		//t.Run("with file, error", func(t *testing.T) {
		//		cmd := NewConfigSetSpecificsCommand(gctx)
		//		cmd.SetArgs([]string{"xxx", "xxx", "xxx"}) // This will fail, but that is ok.
		//		err := cmd.Execute()
		//		// TODO is should really be this: assert.ErrorContains(t, err, "Not found")
		//		assert.ErrorContains(t, err, "Could not load configuration provider")
		//	})
	}
}
