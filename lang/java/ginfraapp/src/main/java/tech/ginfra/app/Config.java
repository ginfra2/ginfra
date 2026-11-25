/*
Package tech.ginfra
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Ginfra helper application.

*/
package tech.ginfra.app;

import picocli.CommandLine;
import tech.ginfra.api.service_giconfig.provider.Provider;
import tech.ginfra.liveconfig.ServiceConfig;
import tech.ginfra.staticconfig.OpenTelemetry;

import java.io.BufferedWriter;
import java.io.FileWriter;
import java.util.Map;

public class Config {

    // --  NAMES -------------------------------------------------------------------------------

    // --  GLOBAL INSTANCE ---------------------------------------------------------------------

    // --  MEMBERS ---------------------------------------------------------------------

    // --  METHODS ---------------------------------------------------------------------

    @CommandLine.Command(name = "props",
            description = "Output known configuration properties.  It will use service_config.json from the CWD.")
    static class ConfigPropsCommand implements Runnable {

        @Override
        public void run() {
            try {
                ServiceConfig sc = ServiceConfig.LoadServiceConfig("service_config.json");

                // We will need a config provider.
                Provider provider = Provider.getProvider(sc.configUri);

                environmentConfig(sc, provider);

            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }
    }

    private static Map<String, String> environmentConfig(ServiceConfig sc, Provider provider) throws Exception {

        // Just this for now
        var env = OpenTelemetry.configOpenTelemetry(provider, sc.deploymentName, sc.name);

        // Write ENV to a file
        BufferedWriter ow = new BufferedWriter(new FileWriter("ENV"));
        try {
            for (String k : env.keySet()) {
                ow.write(k);
                ow.write('=');
                ow.write(env.get(k));
                ow.newLine();
            }
        } finally {
            ow.close();
        }

        return env;
    }


}
