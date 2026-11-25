/*
Package main
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Service entrypoint.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package main

import (
	"<<<Module>>>/local"
	"<<<Module>>>/service"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {

	var (
		c   *local.GContext
		err error
	)

	// Make sure we always print a panic.
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC: %v\n", r)
			panic(r)
		}
	}()

	// Get the command values.
	if err = local.RootCmd.Execute(); err != nil {
		panic(fmt.Sprintf("Bad command: %v", err))
	}

	// Get context and setup
	if c, err = local.GetContext(); err != nil || c == nil {
		panic(fmt.Sprintf("Failed to configure service context: %v", err))
	}
	s := service.Setup(c)

	// Run it
	err = service.RunService(s)
	if err != http.ErrServerClosed {
		fmt.Printf("Server exited with an error: %v", err)
		if local.Logger != nil {
			local.Logger.Fatal("Server exited with an error.", zap.Error(err))
		}
		os.Exit(1)
	}
	os.Exit(0)
}
