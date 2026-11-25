//go:build windows

/*
Package run
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Nix only process decorator.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/

package run

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strconv"
)

// DecorateCommand modifies the given exec.Cmd instance using the specified GContext for customization or configuration.
// The windows implementation does not actually make any modifications.
func DecorateCommand(gctx *app.GContext, c *exec.Cmd) {
	// NOP
}

// TerminateCommand forcibly terminates a running command and its child processes using the TASKKILL command on Windows.
func TerminateCommand(gctx *app.GContext, c *exec.Cmd) {
	gctx.GetLogger().Infow("Killing command.", zap.String(base.LM_COMMAND, c.String()))

	// There is no easy way to do this, so just call the stupid command.
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(c.Process.Pid))
	if e := kill.Run(); e != nil {
		gctx.GetLogger().Warn("May not have kill process Windows process.",
			zap.String(base.LM_PID, strconv.Itoa(c.Process.Pid)))
	}
}

// WaitKilled blocks indefinitely while logging a message, simulating a sleep or wait operation, using the provided GContext.
func WaitKilled(gctx *app.GContext) {
	gctx.GetLogger().Infow("...sleepy time.")
	select {}
}

// SetupINTTrap sets up an interrupt handler to gracefully terminate tasks and exit the program upon interruption.  On
// windows interrupt signals are far rarer but still possible.
func SetupINTTrap(gctx *app.GContext, sctx *ScriptContext) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		gctx.GetLogger().Info("INTERRUPTED.  Killing commands.")
		terminateDeferred(gctx, sctx)
		gctx.GetLogger().Info("Exiting.")
		os.Exit(0)
	}()
}
