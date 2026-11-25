/*
Package app
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Common tools and utilities.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package app

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func init() {
	common.SetupTest()
}

func testRunCommandInContext(gctx *GContext, win, unix string) error {
	var c exec.Cmd
	if runtime.GOOS == "windows" {
		c.Path = win
		c.Args = []string{"."}
	} else {
		c.Path = unix
		c.Args = []string{"."}
	}
	return RunCommandInContext(gctx, &c, "testing")
}

func TestRunCommand(t *testing.T) {
	var gctx GContext
	gctx.Debugging = true

	rctx := GetRootContext()
	gctx.logger = rctx.logger

	t.Run("Command succeed", func(t *testing.T) {
		err := testRunCommandInContext(&gctx, filepath.Join(base.GetGinfraHome(), "test", "local", "nop.cmd"),
			filepath.Join(base.GetGinfraHome(), "test", "local", "nop.sh"))
		assert.NoError(t, err)
	})

	t.Run("Command fail", func(t *testing.T) {
		err := testRunCommandInContext(&gctx, "dir", "ls")
		assert.Error(t, err)
	})
}

// TODO PAss in the function
func testRunCommandInContextLive(gctx *GContext, win, unix string) error {
	var c exec.Cmd
	if runtime.GOOS == "windows" {
		c.Path = win
		c.Args = []string{"."}
	} else {
		c.Path = unix
		c.Args = []string{"."}
	}
	return RunCommandInContextLive(gctx, &c, "testing")
}

func runCommandLivePanic(gctx *GContext, win, unix string) (p bool) {
	p = false
	defer func() {
		if r := recover(); r != nil {
			p = true
		}
	}()

	_ = testRunCommandInContextLive(gctx, win, unix)
	return false
}

func TestRunCommandLive(t *testing.T) {
	var gctx GContext
	gctx.Debugging = true

	rctx := GetRootContext()
	gctx.logger = rctx.logger

	t.Run("Command fail", func(t *testing.T) {
		gctx.ServiceHomePath = base.GetGinfraTmpTestDir()
		err := os.MkdirAll(filepath.Join(base.GetGinfraTmpTestDir(), "log"), os.ModePerm)
		if err != nil {
			panic(err)
		}

		err = testRunCommandInContextLive(&gctx, "dir", "ls")
		assert.Error(t, err)
	})

	t.Run("Command fail no log dir", func(t *testing.T) {
		gctx.ServiceHomePath = base.GetGinfraHome()

		panic := runCommandLivePanic(&gctx, filepath.Join(base.GetGinfraHome(), "test", "local", "nop.cmd"),
			filepath.Join(base.GetGinfraHome(), "test", "local", "nop.sh"))
		assert.Equal(t, true, panic)
	})
}
