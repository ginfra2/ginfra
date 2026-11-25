/*
Package tech.ginfra
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Entrypoint for app.

*/
package tech.ginfra;

import picocli.CommandLine;
import tech.ginfra.app.GinfraApp;

public class Main {

    public static void main(String[] args) {
        CommandLine cmd = new CommandLine(new GinfraApp.RootCommand());
        cmd.setExecutionStrategy(new CommandLine.RunAll());
        cmd.execute(args);
        if (args.length == 0) { cmd.usage(System.out); }
    }
}
