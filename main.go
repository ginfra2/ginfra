/*
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

This is the entrypoint for the ginfra application.  Run the app for help.
*/
package main

import (
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/cmd"
	"gitlab.com/ginfra/ginfra/common"
	"os"
)

func main() {
	// wtf, go.  Pretty much ever other language takes main return as the exit code.
	os.Exit(Main(false))
}

func Main(suppress bool) int {
	e := 0

	common.Setup() // Setup panics if it fails.

	cmd.RootCmd.SilenceUsage = suppress
	err := cmd.Execute()

	if err != nil {
		e = app.ExitErrorCode
	}

	return e
}
