/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Ginterface command.
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	ginterfaces "gitlab.com/ginfra/ginfra/app/ginterface"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/config"
	"gopkg.in/yaml.v3"
)

func init() {
	RootCmd.AddCommand(ginterfaceCmd)
	ginterfaceCmd.AddCommand(NewGinterfaceMergeCmd(app.GetRootContext()), NewGinterfaceStubsCmd(app.GetRootContext()))

	ginterfaceCmd.PersistentFlags().StringVar(&app.GetRootContext().InterfaceDir, config.CL_INTERFACEDIR,
		"", "Interface directory.  If blank, it won't be used.")

	ginterfaceCmd.PersistentFlags().BoolVar(&app.GetRootContext().NoControl, config.CL_NO_CONTROL,
		false, "Do not add control interface to merged interface.")

}

// interfaceCmd represents the ginterface command
var ginterfaceCmd = &cobra.Command{
	Use:   "ginterface",
	Short: "Interface management",
	Long:  `Interface management.`,
}

func NewGinterfaceMergeCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "merge [type] [service name] [Target yaml file names...]",
		Short: "Resolve and merge interfaces into a single file",
		Long: `Resolve and merge interfaces into a single file.  [type] is 'openapi' as it is the only interface
definition supported at this time.  The [Target yaml file names...] is a single or list of files to merge.  Any references
to other files will also be merged.  The files can be in one of three places: 1) the value passed to the --interfacedir
parameter (if given), 2) the current working directory or 3) the Ginfra application ginterfaces/ directory.  They
will be searched in that order.  Files will be processed in reverse order from last file specified on the command line
to the deepest nested reference found.  Some of the merges are destructive, so for instance, the info: section you
want in the final merge should be in the last file on the command line.  The output will be printed to stdout unless a
file is given in the --target parameter.  The overwrite log will be written to stderr.  The interface 'ginfra.yaml' is
automatically merged with everything.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				overwrite []string
				err       error
				i         map[string]interface{}
			)
			switch strings.ToLower(args[0]) {
			case "openapi":
            case "openapi5":
				if gctx.Target == "" {
					i, overwrite, err = ginterfaces.OpenApiLoadSpecs(gctx.InterfaceDir, args[1], gctx.NoControl, args[2:])
					if err == nil {
						if b, errr := yaml.Marshal(i); errr == nil {
							fmt.Println(string(b))
						} else {
							panic(errr) //Extremely unusual for this to happen.
						}
					}

				} else {
					overwrite, err = ginterfaces.OpenApiLoadSaveSpecs(gctx.InterfaceDir, args[1], gctx.Target,
						gctx.NoControl, args[2:])
				}

				if err == nil {
					l := log.New(os.Stderr, "", 0)
					for _, i := range overwrite {
						l.Println("OVERWRITE: " + i)
					}
				}

			default:
				err = base.NewGinfraErrorA("Unknown merge type.", base.LM_TYPE, args[0])
			}
			return err
		},
		Args: cobra.MinimumNArgs(3),
	}
}

func NewGinterfaceStubsCmd(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "stubs [type] [service name] [name] [Target yaml file names...]",
		Short: "Merges interfaces into a single interface and generates code stubs for it.",
		Long: `Resolve and merge interfaces into a single file and generates code studs for it.  It must be run from the
root directory of the service that will use the stubs or the --target parameter points to the root directory.  [type] 
is 'openapi' and 'openapi5' are the only interface definition supported at this time.  They both support server running 
Echo (https://echo.labstack.com/).  The only difference is 'openapi5' supports Echo v5, whereas 'openapi' supports v4.  
The [name] is the name of the interface and will be module name for the stubs.  The [Target yaml file names...] is a single 
or list of files to merge.  Any references to other files will also be merged.  The files can be in one of three places: 1) 
the value passed to the --interfacedir parameter (if given), 2) the current working directory or 3) the Ginfra application 
ginterfaces/ directory. They will be searched in that order.  Files will be processed in reverse order from last file specified 
on the command line to the deepest nested reference found.  Some of the merges are destructive, so for instance, the 
info: section you want in the final merge should be in the last file on the command line.  The stubs will be put in an 
api/ directory.  The interface 'ginfra.yaml' is automatically merged with everything.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				overwrite []string
				err       error
			)
			switch strings.ToLower(args[0]) {
			case "openapi":
				overwrite, err = ginterfaces.OpenApiGenerateStubs(gctx, false, args[1], args[2], gctx.NoControl, args[3:])

				if err == nil {
					for _, i := range overwrite {
						fmt.Println("OVERWRITE: " + i)
					}
				}

			case "openapi5":
				overwrite, err = ginterfaces.OpenApiGenerateStubs(gctx, true, args[1], args[2], gctx.NoControl, args[3:])

				if err == nil {
					for _, i := range overwrite {
						fmt.Println("OVERWRITE: " + i)
					}
				}

			default:
				err = base.NewGinfraErrorA("Unknown stub type.", base.LM_TYPE, args[0])
			}
			return err
		},
		Args: cobra.MinimumNArgs(4),
	}
}
