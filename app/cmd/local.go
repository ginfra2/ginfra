/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage local execution commands
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage/local"
	"gitlab.com/ginfra/ginfra/common/config"
	"os"
	"strconv"
)

var localCmd = NewLocalCommand(app.GetRootContext())

func NewLocalCommand(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:          "local",
		Short:        "Management tools",
		Long:         `Instance management.`,
		SilenceUsage: true,
	}
}

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("Could not get hostname from OS.  %v", err)) // This should never happen.
	}

	RootCmd.AddCommand(localCmd)
	localCmd.AddCommand(NewLocalStart(app.GetRootContext()), NewLocalInitData(app.GetRootContext()))

	localCmd.PersistentFlags().StringVar(&app.GetRootContext().PortalHostnameUrl, config.CL_THIS,
		hostname, "This systems hostname (or ip)")
	localCmd.PersistentFlags().StringVar(&app.GetRootContext().ServiceClass, config.CL_CLASS,
		"", "Service class.  This is the name used for service discovery")

	localCmd.PersistentFlags().IntVarP(&app.GetRootContext().Port, "port",
		"p", 0, // The start command will pick one.
		"Service port.  The default is "+strconv.Itoa(config.DEFAULT_SERVICE_PORT)+".")
	localCmd.PersistentFlags().StringVar(&app.GetRootContext().Specifics, config.CL_SPECIFICS,
		"", "Service specifics.  '[{name,value},{name2,value1}]' json array of name/value pairs.")
	localCmd.PersistentFlags().StringVar(&app.GetRootContext().HostDirPath, config.CL_HOSTDIR,
		".", "Path to the host directory.  'default' is the CWD.")

	localCmd.PersistentFlags().BoolVar(&app.GetRootContext().Clear, config.CL_CLEAR,
		false, "DANGER: tell the service to clear its data stores.")
	localCmd.PersistentFlags().StringVar(&app.GetRootContext().ServiceHomePath, config.CL_SERVICE_HOME,
		"", "path to service repo home")
	localCmd.PersistentFlags().BoolVar(&app.GetRootContext().DontStart, config.CL_DONT_START,
		false, "Don't start the service, but complete registration and other pre-start activities.  Useful for testing.")

}

// #####################################################################################################################
// # GINFRA LOCAL

func NewLocalStart(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "start [config provider uri] [optional: name/value pairs for service specifics (space separated)]",
		Short: "Start a service as a local process.",
		Long: `Start a service as a process on the local machine.  It must be run from the root source directory or the 
service source.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var sp = []string{}
			if len(args) > 1 {
				sp = args[1:]
			}
			if gctx.Port == 0 {
				// Local cannot autoselect a port so give it a default.
				gctx.Port = config.DEFAULT_SERVICE_PORT
			}

			return local.StartLocal(gctx, args[0], sp)
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewLocalInitData(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "initdata [config provider uri] [service name] [init name]",
		Short: "Init a node.",
		Long:  `It will use the init datasource endpoint.  What the endpoint actually does is up to the service/server.  `,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return local.LocalInitData(gctx, args[0], args[1], args[2])
		},
		Args:         cobra.MinimumNArgs(3),
		SilenceUsage: true,
	}
}
