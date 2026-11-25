/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Tool commands.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/tool"
	"gitlab.com/ginfra/ginfra/base"
	"strconv"
	"time"
)

func init() {
	seleniumCmd.AddCommand(NewSeleniumRunCommand(), NewSeleniumStatusCommand(), NewSeleniumLogCommand(),
		NewSeleniumResultCommand(), NewSeleniumStopCommand())
	testCmd.AddCommand(NewTestRunCommand(), NewTestStatusCommand(), NewTestStopCommand(), NewTestAllCommand(),
		NewTestLogCommand(), NewTestRemoveCommand())
	toolCmd.AddCommand(NewInfoCommand(), NewSnoozeCommand(), seleniumCmd, testCmd)
	RootCmd.AddCommand(toolCmd)
}

var toolCmd = NewToolCommand()

func NewToolCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "tool",
		Short:        "General tools",
		Long:         `General tools.`,
		SilenceUsage: true,
	}
}

// =====================================================================================================================
// == INFO

func NewInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "General information.",
		Long:  `Get general information, which isn't much at this time.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.Info(app.GetRootContext())
		},
	}
}

// =====================================================================================================================
// == SNOOZE

func NewSnoozeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "snooze [milliseconds]",
		Short: "Snooze for so many milliseconds.",
		Long:  `Snooze for so many milliseconds.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := strconv.Atoi(args[0])
			if err != nil {
				return base.NewGinfraErrorChild("Bad snooze value.", err)
			}
			time.Sleep(time.Duration(t) * time.Millisecond)
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}
}

// =====================================================================================================================
// == SELENIUM

var seleniumCmd = NewSeleniumCommand()

func NewSeleniumCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "selenium",
		Short:        "Selenium host client.",
		Long:         `Selenium host client.`,
		SilenceUsage: true,
	}
}

func NewSeleniumRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run [host] [target] [path to snippet]",
		Short: "Send snippet to selenium host.",
		Long:  `Send snippet to selenium host.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.SeleniumRun(app.GetRootContext(), args[0], args[1], args[2])
		},
		Args:         cobra.MinimumNArgs(3),
		SilenceUsage: true,
	}
}

func NewSeleniumStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status [host] ",
		Short: "Get status from selenium host.",
		Long:  `Get status from selenium host.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.SeleniumStatus(app.GetRootContext(), args[0])
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewSeleniumLogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "log [host] [id]",
		Short: "Get log from selenium host for the given id.",
		Long:  `Get log from selenium host for the given id.  You can get the id from the status command.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.SeleniumLog(app.GetRootContext(), args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewSeleniumResultCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "result [host] [id]",
		Short: "Get test result from selenium host for the given id.",
		Long:  `Get test result  from selenium host for the given id.  You can get the id from the status command.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.SeleniumResult(app.GetRootContext(), args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewSeleniumStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [host] [id] ",
		Short: "Get stop a snippet on the host.",
		Long:  `Get stop a snippet on the host.  The id is ignored in this version, since only one snippet may be running.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.SeleniumStop(app.GetRootContext(), args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

// =====================================================================================================================
// == TEST

var testCmd = NewTestCommand()

func NewTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "test",
		Short:        "Test tools",
		Long:         `Test tools.`,
		SilenceUsage: true,
	}
}

func NewTestRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run [uid] [target] [path to test]",
		Short: "Send test to test host.",
		Long: `Send test to test host.  Uid is any string that identifies the user and is not checked for this cli tool.
Deployment, service name, service host and authorization token is required.  All four can be passed on the command line, 
but if any are not, the command will try to derive them from the Hosting.json file in the current working directory and 
any associated configuration service.  If the Hosting.json file does not exist, it will assume the deployment is 'default'. 
It will assume the service is 'test' unless it is passed.  If the authorization token is not passed and it cannot be found in 
a configuration service, the command will fail.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.TestRun(app.GetRootContext(), args[0], args[1], args[2])
		},
		Args:         cobra.MinimumNArgs(3),
		SilenceUsage: true,
	}
}

func NewTestStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [uid] [id]",
		Short: "Stop a test on the test host for a given id.",
		Long: `Stop a test on the test host for a given id.  Uid is any string that identifies the user and is not checked for this cli tool. 
The stop may not happen instantly and status could be checked until it is stopped.
Deployment, service name, service host and authorization token is required.  All four can be passed on the command line, 
but if any are not, the command will try to derive them from the Hosting.json file in the current working directory and 
any associated configuration service.  If the Hosting.json file does not exist, it will assume the deployment is 'default'. 
It will assume the service is 'test' unless it is passed.  If the authorization token is not passed and it cannot be found in 
a configuration service, the command will fail.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.TestStop(app.GetRootContext(), args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewTestStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status [uid] [id]",
		Short: "Get status from test host for a given id.",
		Long: `Get status from test host for a given id.  Uid is any string that identifies the user and is not checked for this cli tool. 
Deployment, service name, service host and authorization token is required.  All four can be passed on the command line, 
but if any are not, the command will try to derive them from the Hosting.json file in the current working directory and 
any associated configuration service.  If the Hosting.json file does not exist, it will assume the deployment is 'default'. 
It will assume the service is 'test' unless it is passed.  If the authorization token is not passed and it cannot be found in 
a configuration service, the command will fail.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.TestStatus(app.GetRootContext(), args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewTestAllCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Get status from test host for all tasks.",
		Long: `Get status from test host for all tasks.   Uid is any string that identifies the user and is not checked for this cli tool.
Deployment, service name, service host and authorization token is required.  All four can be passed on the command line, 
but if any are not, the command will try to derive them from the Hosting.json file in the current working directory and 
any associated configuration service.  If the Hosting.json file does not exist, it will assume the deployment is 'default'. 
It will assume the service is 'test' unless it is passed.  If the authorization token is not passed and it cannot be found in 
a configuration service, the command will fail.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.TestAll(app.GetRootContext(), args[0])
		},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
}

func NewTestLogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "log [id]",
		Short: "Get log from test host for a given id.",
		Long: `Get log from test host for a given id.  Uid is any string that identifies the user and is not checked for this cli tool.
Deployment, service name, service host and authorization token is required.  All four can be passed on the command line, 
but if any are not, the command will try to derive them from the Hosting.json file in the current working directory and 
any associated configuration service.  If the Hosting.json file does not exist, it will assume the deployment is 'default'. 
It will assume the service is 'test' unless it is passed.  If the authorization token is not passed and it cannot be found in 
a configuration service, the command will fail.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.TestLog(app.GetRootContext(), args[0], args[1])
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}

func NewTestRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [ids...]",
		Short: "Move the given tests by id.",
		Long: `Move the given tests by id.   Uid is any string that identifies the user and is not checked for this cli tool.
Deployment, service name, service host and authorization token is required.  All four can be passed on the command line, 
but if any are not, the command will try to derive them from the Hosting.json file in the current working directory and 
any associated configuration service.  If the Hosting.json file does not exist, it will assume the deployment is 'default'. 
It will assume the service is 'test' unless it is passed.  If the authorization token is not passed and it cannot be found in 
a configuration service, the command will fail.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootSetup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tool.TestRemove(app.GetRootContext(), args[0], args[1:]...)
		},
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
	}
}
