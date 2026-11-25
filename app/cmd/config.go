/*
Package cmd
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Configuration related commands.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/configctl"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
)

func init() {

	configSetCommand.AddCommand(NewConfigSetSpecificsCommand(app.GetRootContext()))
	configCmd.AddCommand(configSetCommand)
	RootCmd.AddCommand(configCmd)
}

var configCmd = NewConfigCommand()

func NewConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "config",
		Short:        "Config management",
		Long:         `Config management.`,
		SilenceUsage: true,
	}
}

// =====================================================================================================================
// == SET

var configSetCommand = NewConfigSetCommand()

func NewConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Set Config.",
		Long:  `Set Config.`,
	}
}

func NewConfigSetSpecificsCommand(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "specifics [service] [name] [value]",
		Short: "Set a service specifics config value.",
		Long: `Set a service specifics config value.  
Deployment, service name and configuration auth token is required.  It will attempt to get deployment and auth from a
Hosting.json file in the cwd.  This command isn't of much use for local service deployment, since the user will have 
complete control over the configuration provider'`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			hf, err := manage.LoadHostFileCwd()
			if err != nil {
				return base.NewGinfraErrorChild("Hosting.json file not found.",
					err)
			}

			return configctl.SetSpecifics(gctx, hf.ConfigPortalHostUrl, hf.ConfigAccessToken, hf.DeploymentName, args[0],
				args[1], args[2])
		},
		Args: cobra.MinimumNArgs(3),
	}
}
