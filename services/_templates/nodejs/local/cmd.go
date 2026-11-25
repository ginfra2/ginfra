/*
Package local
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt if unaltered.  You are welcome to alter and apply any
license you wish to this file.

Local command processor.  This is the bridge between the command line arguments and a configured context.
*/
package local

import (
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
)

var (
	clr   bool
	rpath string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "",
	Short: "Service start",
	Long: `Local command service.  It accepts a single argument that is the root path to the service.  
In that path it expects to find a configuration file named "service_config.json".`,
	Run: func(cmd *cobra.Command, args []string) {
		rpath = args[0]
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&clr, config.CL_CLEAR, false, "DANGER: clr data stores and start fresh.")
}

func (gctx *GContext) GetCmdContext() error {
	var err error

	gctx.Scfg, err = shared.LoadServiceConfig(rpath)
	if err != nil {
		return err
	}
	gctx.Scfg.ServiceHome = rpath

	return err
}

// ^^###################################################################################################################
// # ADDITIONAL CODE GOES BELOW HERE
