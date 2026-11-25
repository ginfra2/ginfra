/*
Package main
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

This is the entrypoint for the test stub.
*/
package main

import (
	"gitlab.com/ginfra/ginfra/test/ginfra_stub/cmd"
	"os"
)

func main() {
	e := 0

	// Setup panics if it fails.
	err := cmd.Execute()

	if err != nil {
		e = 1
	}

	os.Exit(e)
}
