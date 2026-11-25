/*
Package app
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Common tools and utilities.
*/
package app

import (
	"fmt"
	"gitlab.com/ginfra/ginfra/base"
	"golang.org/x/mod/modfile"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ServiceSpec struct {
	Name       string `json:"name"`
	Class      string `json:"class"`
	ModuleName string `json:"module_name"`
}

// RunCommandInContext executes a given command within the provided GContext, capturing stdout and stderr for debugging.
// If execution fails, it wraps the error with additional context information including command output and annotations.
// Logging is conditional on the context's Debugging setting; debug output is provided if enabled.
func RunCommandInContext(context *GContext, cmd *exec.Cmd, note string) error {
	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err == nil {
		if context.Debugging {
			context.GetLogger().Debugw(fmt.Sprintf("Output from %s.", note), "stdout",
				stdoutBuf.String())
		}
	} else {
		err = base.NewGinfraErrorChildA("Cmd failed.", err, base.LM_STDIN, stdoutBuf.String(), base.LM_STDOUT,
			stderrBuf.String(), base.LM_COMMAND, cmd.String())
	}
	return err
}

// RunCommandInContextLive runs a command within the given context, writing its output to a log file in the service's
// log directory. Returns an error if the command fails or if there is an issue opening the log file.
func RunCommandInContextLive(context *GContext, cmd *exec.Cmd, note string) error {
	lp := filepath.Join(context.ServiceHomePath, base.LMCONFIG_LOGDIR, "std.log")
	f, err := os.OpenFile(lp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(base.NewGinfraErrorChildA("Error opening std log file.  Does the service have a log/ directory?",
			err, base.LM_FILEPATH, lp))
	}
	defer func() {
		_ = f.Close()
	}()
	cmd.Stderr = f
	cmd.Stdout = f

	err = cmd.Run()
	if err != nil {
		err = base.NewGinfraErrorChildA("Cmd failed.", err, base.LM_COMMAND, cmd.String())
	}
	return err
}

// GetFullModule retrieves the module name from a `go.mod` file in the specified target directory.
// It returns the module name as a string and an error if the `go.mod` file is not found or cannot be parsed.
func GetFullModule(targetDirectory string) (string, error) {
	var (
		err        error = nil
		moduleName string
	)
	modFp := filepath.Join(targetDirectory, "go.mod")

	if _, err = os.Stat(modFp); err == nil {
		ast := &modfile.File{}

		var raw []byte
		if raw, err = os.ReadFile(modFp); err == nil {
			if ast, err = modfile.Parse("go.mod", raw, nil); err == nil {
				moduleName = ast.Module.Mod.Path
			}
		}
	}

	return moduleName, err
}
