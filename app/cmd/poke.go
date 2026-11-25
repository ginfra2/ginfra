/*
Package cmd
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Poke commands.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/poke"
	"gitlab.com/ginfra/ginfra/base"
)

func init() {
	pokeCmd.AddCommand(NewPokePostCommand(app.GetRootContext()), NewPokeGetCommand(app.GetRootContext()))
	RootCmd.AddCommand(pokeCmd)
}

var pokeCmd = NewPokeCommand()

func NewPokeCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "poke",
		Short:        "Poke commands",
		Long:         `Useful for interacting with REST service and like endpoints.`,
		SilenceUsage: true,
	}
}

// =====================================================================================================================
// == POST

func NewPokePostCommand(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "post [url] [auth token] [name/value pairs sent as json data]...",
		Short: "Poke a POST endpoint.",
		Long:  `Poke a POST endpoint.  Data will be sent as JSON.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args)%2 != 0 {
				return base.NewGinfraErrorChild("Odd number of name/value arguments.", nil)
			}
			var (
				rep string
				err error
			)
			if len(args) > 2 {
				rep, err = poke.Post(app.GetRootContext(), args[0], args[1], args[2:])
			} else {
				rep, err = poke.Post(app.GetRootContext(), args[0], args[1], []string{})
			}
			if err != nil {
				return err
			}
			gctx.GetLogger().Infof(rep)
			return nil
		},
		Args: cobra.MinimumNArgs(2),
	}
}

func NewPokeGetCommand(gctx *app.GContext) *cobra.Command {
	return &cobra.Command{
		Use:   "get [url] [auth token] [name/value pairs sent as json data]...",
		Short: "Poke a GET endpoint.",
		Long:  `Poke a GET endpoint.  Data will be sent as JSON.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args)%2 != 0 {
				return base.NewGinfraErrorChild("Odd number of name/value arguments.", nil)
			}
			var (
				rep string
				err error
			)
			if len(args) > 2 {
				rep, err = poke.Get(app.GetRootContext(), args[0], args[1], args[2:])
			} else {
				rep, err = poke.Get(app.GetRootContext(), args[0], args[1], []string{})
			}
			if err != nil {
				return err
			}
			gctx.GetLogger().Infof(rep)
			return nil
		},
		Args: cobra.MinimumNArgs(2),
	}
}
