/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Generation tool commands.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/generate"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/config"
)

// interfaceCmd represents the ginterface command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generation tools",
	Long:  `Interface management.`,
}

func NewGenerateService() *cobra.Command {
	return &cobra.Command{
		Use:   "service [Service name] [Service type] [Module Name]",
		Short: "Create the code scaffold for a service.",
		Long: `The following service classes are supported.  All generated files will be put in the current working directory.
It will look for the template for the service in a subdirectory named [Service type] in the templates directory.  
The default templates directory will be the templates provided by the Ginfra application.  However you can point to another
directory with the --templates parameter.  To get a list of the templates provided by the Ginfra app run the 
"ginfra generate list" command.  The services provided by the Ginfra app will come with a giconfig client and server_control
api stubs.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := app.GetRootContext().ServicePath
			if path == "" {
				path = args[0]
			}

			sclass := app.GetRootContext().ServiceClass
			if sclass == "" {
				sclass = args[0]
			}

			td := app.GetRootContext().Source
			if td == "" {
				td = base.GetGinfraTemplatesDir()
			}

			err := generate.GenerateService(app.GetRootContext(), args[0], sclass, path, args[1], args[2], td)

			if err == nil {
				fmt.Println("IMPORTANT: Generate assumes the GINFRA_HOME environment variable points to the ginfra repo and it is consistent")
				fmt.Println("with where your services (including ginfra/gservices) repos are.  If you get build errors, you may need to ")
				fmt.Println("adjust this in your service's go.mod file.   Look for the following line:")
				fmt.Println("        replace gitlab.com/ginfra/ginfra vX.X.X =>")
				fmt.Println("and adjust the last to the right of => to point to the correct location.  I recommend using a relative path,")
				fmt.Println("particularly if you checkout all repos to the same directory.")
				fmt.Println("")
				fmt.Println("IMPORTANT: You should run 'task stubs' to make sure any APIs are correct for your newly generated service.")
				fmt.Println("even if you are not adding any new APIS.")
			}

			return err

		},
		Args: cobra.MinimumNArgs(3),
	}
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Generation tools",
	Long:  `Generation management.`,
}

func NewUpdateService() *cobra.Command {
	return &cobra.Command{
		Use:   "service",
		Short: "Update the code scaffold for a service.",
		Long: `This must be run in the root directory of the service you wish to update.  It will look for the 
template for the service in a subdirectory named the service type in the templates directory.  The service type is fixed when
the service is originally generated.  The default templates directory will be the templates provided by the Ginfra 
application.  However you can point to another directory with the --templates parameter.  IMPORTANT!!!!  Run 'go mod tidy' after
this command or you will likely get compiler errors.  Also, if you are using api generator, you will need to generate the
apis again.  Unless you did something custom, running 'task stubs' should do this.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			td := app.GetRootContext().Source
			if td == "" {
				td = base.GetGinfraTemplatesDir()
			}
			return generate.UpdateService(app.GetRootContext(), td)
		},
	}
}

func NewGenerateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List provided service templates",
		Long:  `List service templates provided by the Ginfra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("[Service type] : Description")
			fmt.Println("apiecho        : Echo server (https://echo.labstack.com/) for servicing OpenAPI requests using Go.")
			fmt.Println("server         : General service for hosting a third party server (such as MongoDb or Postgres).")
			fmt.Println("micronaut      : Micronaut server (https://micronaut.io/) for servicing OpenAPI requests using Java.")
			fmt.Println("vuevitets      : Vue.js/Vite service (https://vuejs.org/) for creating a frontend application for Node.js using TypeScript.")
			fmt.Println("nodejs         : NodeJS server (https://nodejs.org/) for creating server side applications using JavaScript.")
		},
	}
}

func init() {
	RootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(NewGenerateService(), NewGenerateListCmd())
	RootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(NewUpdateService())

	generateCmd.PersistentFlags().StringVar(&app.GetRootContext().Source, config.CL_TEMPLATES,
		"", "Service path.  This will be used for things like docker container name.  The default is the service name.")
	generateCmd.PersistentFlags().StringVar(&app.GetRootContext().ServicePath, config.CL_PATH,
		"", "Service path.  This will be used for things like docker container name.  The default is the service name.")

}
