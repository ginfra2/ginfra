/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage hosted commands.
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage/host"
	"gitlab.com/ginfra/ginfra/common/config"
)

var hostCmd = NewHostCommand(app.GetRootContext())

func NewHostCommand(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:          "host",
		Short:        "Host management tools",
		Long:         `Host management for container based deployments.`,
		SilenceUsage: true,
	}
}

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("Could not get hostname from OS.  %v", err)) // This should never happen.
	}

	RootCmd.AddCommand(hostCmd)
	hostCmd.AddCommand(NewManageHostInitCmd(app.GetRootContext()), NewManageHostStartCmd(app.GetRootContext()),
		NewManageHostStopCmd(app.GetRootContext()), NewManageHostListCmd(app.GetRootContext()),
		NewManageHostShowCmd(app.GetRootContext()), NewManageHostDestroyCmd(app.GetRootContext()),
		NewManageHostStoreCmd(app.GetRootContext()), NewManageHostInitData(app.GetRootContext()),
		NewManageHostResetCmd(app.GetRootContext()), NewManageHostPutCmd(app.GetRootContext()))

	hostCmd.PersistentFlags().StringVar(&app.GetRootContext().PortalHostnameUrl, config.CL_THIS,
		hostname, "Platform hostname (or ip)")
	hostCmd.PersistentFlags().StringVar(&app.GetRootContext().ServiceClass, config.CL_CLASS,
		"", "Service class.  This is the name used for service discovery")

	hostCmd.PersistentFlags().IntVarP(&app.GetRootContext().Port, "port",
		"p", 0, // The start command will pick one.
		"Service port.  The default is "+strconv.Itoa(config.DEFAULT_SERVICE_PORT)+".")
	hostCmd.PersistentFlags().StringVar(&app.GetRootContext().Specifics, config.CL_SPECIFICS,
		"", "Service specifics.  '[{name,value},{name2,value1}]' json array of name/value pairs.")
	hostCmd.PersistentFlags().StringVar(&app.GetRootContext().HostDirPath, config.CL_HOSTDIR,
		".", "Path to the host directory.  'default' is the CWD.")

	hostCmd.PersistentFlags().BoolVar(&app.GetRootContext().NoRollback, config.CL_NO_ROLLBACK,
		false, "Do not rollback containers if there is in error.  It will still rollback any network and configuration.  Useful for testing but it can leave a mess you will have to clean up.")
	hostCmd.PersistentFlags().StringVar(&app.GetRootContext().CommandLine, config.CL_COMMAND_LINE,
		"", "Pass the string as the command line when starting the container.  This should be used for debugging and special situations only.")
	hostCmd.PersistentFlags().IntVar(&app.GetRootContext().Debugger, config.CL_DEBUGGER,
		0, "Attach a debugger to the executable.  While containers can choose their own port for the debugger to listen, the ginfra default is 40001.")

	hostCmd.PersistentFlags().StringVar(&app.GetRootContext().Gpus, config.CL_GPUS,
		"", "GPU specification to be passed to container host (typically with the --gpus command line flag or equivalent json call).")

	hostCmd.PersistentFlags().BoolVar(&app.GetRootContext().DontStart, config.CL_DONT_START,
		false, "Don't start the service, but complete registration and other pre-start activities.  Useful for testing. the target container must support it.")
}

// #####################################################################################################################
// # GINFRA HOST

func NewManageHostInitCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "init [docker image for configuration service]",
		Short: "Initialize a host based deployment.",
		Long: `It will create a file in current working directory (cwd) called 'Hosting.json' which contains metadata for the
deployment.  It will start the given docker image as a configuration provider and test that it is up.  All future management
of this deployment will rely on the Hosting.json file being present in the cwd.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := host.HostInit(gctx, args[0])
			return err
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewManageHostStartCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "start [docker image for the service] [optional: name/value pairs for service specifics (space separated)]",
		Short: "Start a service.",
		Long: `It will look for a file Hosting.json (created by the init command) in the current working directory.  If the
file is not present it will fail.  Limit the command line specifics to three pairs.  If you need more than that then use
the json passed in the --specifics flag.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var sp = []string{}
			if len(args) > 1 {
				sp = args[1:]
			}
			return host.HostStart(gctx, args[0], sp)
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewManageHostListCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List hosting information.",
		Long: `It will look for a file Hosting.json (created by the init command) in the current working directory.  If the
file is not present it will fail.  `,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return host.HostedList(gctx)
		},
		SilenceUsage: true,
	}
}

func NewManageHostShowCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "show [name]",
		Short: "Show service information.",
		Long: `It will look for a file Hosting.json (created by the init command) in the current working directory.  If the
file is not present it will fail.  `,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return host.HostedShow(gctx, args[0])
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewManageHostStopCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [name]",
		Short: "Stop a service.",
		Long: `It will look for a file Hosting.json (created by the init command) in the current working directory.  If the
file is not present it will fail.  `,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return host.HostedStop(gctx, args[0])
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewManageHostDestroyCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "destroy",
		Short: "Destroy a deployment.",
		Long: `It will look for a file Hosting.json (created by the init command) in the current working directory.  If the
file is not present it will fail.  `,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return host.HostedDestroy(gctx)
		},
		SilenceUsage: true,
	}
}

func NewManageHostInitData(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "initdata [service name] [init name]",
		Short: "Init a node.",
		Long:  `It will use the init datasource endpoint.  What the endpoint actually does is up to the service/server.  `,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return host.HostedInitData(gctx, args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewManageHostResetCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "reset [service name]",
		Short: "Reset a service.",
		Long:  `It will use the reset endpoint.  What the endpoint actually does is up to the service/server.  `,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return host.HostedReset(gctx, args[0])
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewManageHostStoreCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "store [name] [path] [deployment] [service]",
		Short: "Store a file in the configuration service.",
		Long: `It will store a file found at [path] with the name of [name].  [deployment] and [service] may be empty and they
are not dependent on the deployment and service actually running.  The stored file will persist for the life of the 
configuration server.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			d := ""
			if len(args) > 2 {
				d = args[2]
			}
			s := ""
			if len(args) > 3 {
				s = args[3]
			}
			return host.HostedStore(gctx, args[0], args[1], d, s)
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewManageHostPutCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "put [name] [value]",
		Short: "Put a value in the configuration service`.",
		Long:  `Put a value in the configuration service.  The value is passed as a string.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return host.HostedPut(gctx, args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}
