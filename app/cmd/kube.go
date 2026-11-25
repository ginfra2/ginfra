/*
Package cmd
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes commands.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage/kube"
	"gitlab.com/ginfra/ginfra/common/config"
	"strconv"
)

var kCmd = NewKubeCommand()

func NewKubeCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "kube",
		Short:        "Kubernetes management",
		Long:         `Kubernetes management.`,
		SilenceUsage: true,
	}
}

func init() {

	RootCmd.AddCommand(kCmd)
	kCmd.AddCommand(NewManageKubeInitCmd(app.GetRootContext()), NewManageKubeDestroyCmd(app.GetRootContext()),
		NewManageKubeStartCmd(app.GetRootContext()), NewManageKubeStopCmd(app.GetRootContext()),
		NewManageKubePortalCmd(app.GetRootContext()), NewManageKubeStoreCmd(app.GetRootContext()),
		NewManageKubeInitData(app.GetRootContext()))

	kCmd.PersistentFlags().StringVar(&app.GetRootContext().ServiceClass, config.CL_CLASS,
		"", "Service class.  This is the name used for service discovery")

	kCmd.PersistentFlags().IntVarP(&app.GetRootContext().Port, "port",
		"p", 0, // The start command will pick one.
		"Service port.  The default is "+strconv.Itoa(config.DEFAULT_SERVICE_PORT)+".")
	kCmd.PersistentFlags().StringVar(&app.GetRootContext().Specifics, config.CL_SPECIFICS,
		"", "Service specifics.  '[{name,value},{name2,value1}]' json array of name/value pairs.")
	kCmd.PersistentFlags().StringVar(&app.GetRootContext().HostDirPath, config.CL_HOSTDIR,
		".", "Path to the host directory.  'default' is the CWD.")

	kCmd.PersistentFlags().BoolVar(&app.GetRootContext().Clear, config.CL_CLEAR,
		false, "DANGER: tell the service to clear its data stores.")
	kCmd.PersistentFlags().StringVar(&app.GetRootContext().ServiceHomePath, config.CL_SERVICE_HOME,
		"", "path to service repo home")
	//kCmd.PersistentFlags().BoolVar(&app.GetRootContext().DontStart, config.CL_DONT_START,
	//	false, "Don't start the service, but complete registration and other pre-start activities.  Useful for testing.")

	kCmd.PersistentFlags().StringVar(&app.GetRootContext().PortalHostnameUrl, config.CL_PORTAL,
		"", "Cluster portal address.  Recommend using FQDN.")

	kCmd.PersistentFlags().BoolVar(&app.GetRootContext().NoRollback, config.CL_NO_ROLLBACK,
		false, "Do not rollback containers if there is in error.  It will still rollback any network and configuration.  Useful for testing but it can leave a mess you will have to clean up.")
	//kCmd.PersistentFlags().StringVar(&app.GetRootContext().CommandLine, config.CL_COMMAND_LINE,
	//	"", "Pass the string as the command line when starting the container.  This should be used for debugging and special situations only.")
	//kCmd.PersistentFlags().IntVar(&app.GetRootContext().Debugger, config.CL_DEBUGGER,
	//	0, "Attach a debugger to the executable.  While containers can choose their own port for the debugger to listen, the ginfra default is 40001.")

	kCmd.PersistentFlags().StringVar(&app.GetRootContext().InitContainer, config.CL_INIT_CONTAINER,
		"", "Use an alternate container image for the Init Container.  The default image does nothing more than copy the configuration into the live containers.  Any custom container must run the following command:  cp /initconfig/service_config.json /config/service_config.json")

}

func NewManageKubeInitCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "init [docker image for configuration service]",
		Short: "Initialize a kubernetes based deployment.",
		Long: `It will create a file in current working directory (cwd) called 'Hosting.json' which contains metadata for the
deployment.  It will start the given docker image as a configuration provider and test that it is up.  All future management
of this deployment will rely on the Hosting.json file being present in the cwd.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := kube.KubeHostInit(gctx, args[0])
			return err
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewManageKubeDestroyCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "destroy",
		Short: "Destroy a kubernetes based deployment.",
		Long:  `It will look for the 'Hosting.json' file in the current working directory for information on the deployment.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := kube.KubeHostDestroy(gctx)
			return err
		},
		SilenceUsage: true,
	}
}

func NewManageKubeStartCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "start [name][docker image for service]",
		Short: "Start a service.",
		Long:  `Start a service`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if len(args) > 2 {
				err = kube.KubeStart(gctx, args[0], args[1], args[2:])
			} else {
				err = kube.KubeStart(gctx, args[0], args[1], []string{})
			}

			return err
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewManageKubeStopCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [name]",
		Short: "Stop a service.",
		Long:  `Stop a service`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := kube.KubeStop(gctx, args[0])
			return err
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewManageKubeInitData(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "initdata [service name] [init name]",
		Short: "Init a node.",
		Long:  `It will use the init datasource endpoint.  What the endpoint actually does is up to the service/server.  `,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return kube.KubeInitData(gctx, args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewManageKubePortalCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "portal [name] [address]",
		Short: "Add a portal (ingress).",
		Long:  `Add a portal (ingress).  The default portal is named '*' or you can pass 'default' or an empty string.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := kube.KubePortal(gctx, args[0], args[1])
			return err
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewManageKubeStoreCmd(gctx *app.GContext) *cobra.Command {
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
			return kube.KubeStore(gctx, args[0], args[1], d, s)
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}
