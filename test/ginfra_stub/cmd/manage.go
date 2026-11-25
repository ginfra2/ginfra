/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Commands for the stub.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/test/ginfra_stub/manage"
)

func init() {
	RootCmd.AddCommand(localCmd)
}

// interfaceCmd represents the ginterface command
var localCmd = NewLocalCommand(app.GetRootContext())

func NewLocalCommand(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "local",
		Short: "Run as local process",
		Long: `Run as local process.  It will end the process if a file called 'flagfile.txt' appears in the working
directory.  There is an overall 3 minute timeout regardless of the file, to help keep ghosts away.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return manage.Local(gctx)
		},
		SilenceUsage: true,
	}
}
