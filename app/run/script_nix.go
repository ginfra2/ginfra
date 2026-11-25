//go:build linux || darwin

/*
Package run
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Nix only process decorator.

*/

package run

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

// DecorateCommand modifies the given exec.Cmd to set `SysProcAttr` such that a new process group is created.
func DecorateCommand(gctx *app.GContext, c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// TerminateCommand forcibly terminates a given command and releases associated resources to prevent process hanging.
// It attempts to kill the process group of the provided command and logs any failures encountered during termination.
// The function uses GContext for logging and handles syscall errors, ensuring proper cleanup of detached processes.
func TerminateCommand(gctx *app.GContext, c *exec.Cmd) {
	gctx.GetLogger().Infow("Killing command.", zap.String(base.LM_COMMAND, c.String()))

	if c != nil && c.Process != nil {
		pgid, err := syscall.Getpgid(c.Process.Pid)
		if err == nil {
			// Using the exec calls just aren't reliable on some linux distros.
			if err = syscall.Kill(-pgid, 2); err == nil {
				// err = c.Wait()
				// Using release instead of wait because there isn't anything to wait for.  Any pipe will be abandoned
				// anyway whereas when using Wait there is a chance of one ghosting and hanging the call.
				_ = c.Process.Release()

			} else {
				gctx.GetLogger().Error("Couldn't kill command", zap.String(base.LM_PID, strconv.Itoa(pgid)),
					zap.Error(err))

			}
		}
	}

}

// WaitKilled waits indefinitely for a termination signal, logging a message to indicate the application is idle.
func WaitKilled(gctx *app.GContext) {
	// Just wait for the term signal
	gctx.GetLogger().Infow("...sleepy time.")
	select {}
}

// SetupINTTrap initializes a signal trap for SIGINT and SIGTERM to handle graceful termination of the application.
// It ensures proper cleanup by invoking terminateDeferred and logs interruption and exit events.
// gctx represents the global application context, while sctx provides the script-specific execution context.
func SetupINTTrap(gctx *app.GContext, sctx *ScriptContext) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		gctx.GetLogger().Info("INTERRUPTED.  Killing commands.")
		terminateDeferred(gctx, sctx)
		gctx.GetLogger().Info("Exiting.")
		os.Exit(0)
	}()
}
