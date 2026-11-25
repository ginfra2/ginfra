/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Root cli command.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/common/config"
)

var (
	DebuggingFlag *bool
	BlobFlag      *bool
)

var RootCmd = NewRootCommand(app.GetRootContext())

func NewRootCommand(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "ginfra",
		Short: "GINFRA management tool",
		Long:  `GINFRA management tool`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		Version: "main", // TODO set at build time.
	}
}

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	DebuggingFlag = RootCmd.PersistentFlags().Bool("debugging", false, "turn on debugging log")
	BlobFlag = RootCmd.PersistentFlags().Bool("blob", false, "output as a blob instead of logging.")
	RootCmd.PersistentFlags().StringVar(&app.GetRootContext().Target, "target",
		"", "output target directory or file path")
	RootCmd.PersistentFlags().StringVar(&app.GetRootContext().Deployment, config.CL_DEPLOYMENT,
		"default", "Deployment name.  'default' is the default.")
	RootCmd.PersistentFlags().StringVar(&app.GetRootContext().Service, config.CL_SERVICE,
		"", "Service name.  There is no default.")
	RootCmd.PersistentFlags().StringVar(&app.GetRootContext().Host, config.CL_HOST,
		"", "Target host resolvable name or IP.  There is no default.")
	RootCmd.PersistentFlags().StringVar(&app.GetRootContext().Authtoken, config.CL_AUTH, "",
		"Authorization token, typically used when working with a service.  When starting a new service it will use this as the service auth token.")
	RootCmd.PersistentFlags().StringVar(&app.GetRootContext().Bundle, config.CL_BUNDLE,
		"", "Include the given bundle (typically a zip file) in the operation.  Typically ued to provide crypo/cert files to a service/server.")
}

func rootSetup() error {
	if *DebuggingFlag {
		app.GetRootContext().TurnDebuggingOn()
	}
	app.GetRootContext().Blob = *BlobFlag
	return nil
}
