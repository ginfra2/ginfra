/*
Package tech.ginfra
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Ginfra helper application.

*/
package tech.ginfra.app;

import picocli.CommandLine;

public class GinfraApp {

    // --  NAMES -------------------------------------------------------------------------------

    // --  GLOBAL INSTANCE ---------------------------------------------------------------------

    // --  MEMBERS ---------------------------------------------------------------------

    // --  METHODS ---------------------------------------------------------------------

    @CommandLine.Command(name = "config", subcommands = { Config.ConfigPropsCommand.class,
            CommandLine.HelpCommand.class },
            description = "Configuration helpers.")
    static class ConfigCommand implements Runnable {

        @Override
        public void run() { }
    }

    @CommandLine.Command(name = "", subcommands = { ConfigCommand.class,
            CommandLine.HelpCommand.class })
    static public class RootCommand implements Runnable {

        @Override
        public void run() { }
    }

}
