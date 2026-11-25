/*
Package cmd
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Root cli command.
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	DebuggingFlag *bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "ginfra_stub",
	Short:   "GINFRA stub tool",
	Long:    `GINFRA stub tool`,
	Version: "0.0.5++", // TODO set at build time.
}

func Execute() error {
	return RootCmd.Execute()
}

func init() {
}
