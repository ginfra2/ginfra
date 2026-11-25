/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Run scripts commands.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	DoRun "gitlab.com/ginfra/ginfra/app/run"
)

func init() {
	RootCmd.AddCommand(NewRunCommand(app.GetRootContext()))
}

func NewRunCommand(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "run [run script]",
		Short: "Orchestration runner",
		Long:  `Orchestration runner.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return DoRun.Run(gctx, args[0])
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}
