/*
Package tech.ginfra
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Ginfra class
*/
package tech.ginfra;

import org.slf4j.Logger;

import static net.logstash.logback.argument.StructuredArguments.kv;
import org.slf4j.LoggerFactory;
import tech.ginfra.api.service_giconfig.provider.Provider;
import tech.ginfra.liveconfig.ContainerSpecifics;
import tech.ginfra.liveconfig.ServiceConfig;
import tech.ginfra.liveconfig.ServiceConfigSpecifics;

import java.nio.file.Paths;

public class Ginfra {

    // --  NAMES -------------------------------------------------------------------------------
    public final static String SPECIFICS_FILE_NAME = "specifics.json";
    public final static String SERVICE_CONFIG_FILE_NAME = "service_config.json";

    // --  GLOBAL INSTANCE ---------------------------------------------------------------------

    private static Ginfra ginfra;

    // --  MEMBERS ---------------------------------------------------------------------

    private Provider configProvider;
    private ContainerSpecifics containerSpecifics;
    private ServiceConfigSpecifics serviceSpecifics;
    private ServiceConfig serviceConfig;
    private Logger logPeople;
    private Logger logMachine;
    private Logger logAccess;

    // --  METHODS ---------------------------------------------------------------------

    public static void LoadGinfra(String homepath) throws Exception {

        ginfra = new Ginfra();

        // Prepare loggers.
        ginfra.logPeople = LoggerFactory.getLogger("people");
        ginfra.logMachine = LoggerFactory.getLogger("machine");
        ginfra.logAccess = LoggerFactory.getLogger("access");

        // Load container specifics
        try {
            ginfra.containerSpecifics = ContainerSpecifics.LoadSpecifics(Paths.get(homepath, SPECIFICS_FILE_NAME).toString());
        } catch (Exception e) {
           ginfra.logPeople.info("Container specifics not loaded.", kv("error", e.getMessage()));
           throw e;
        }

        // Load server config.
        var configPath = Paths.get(homepath, SERVICE_CONFIG_FILE_NAME).toString();
        try {
            ginfra.serviceConfig = ServiceConfig.LoadServiceConfig(configPath);
        } catch (Exception e) {
            // This is fatal.
            ginfra.logPeople.error("Service config not loaded.", kv("error", e.getMessage()),
                    kv("path", configPath));
            throw e;
        }

        try {
            ginfra.configProvider = Provider.getProvider(ginfra.serviceConfig.configUri);
            ginfra.configProvider.SetConfigAccessAuth(ginfra.serviceConfig.configToken);
        } catch(Exception e) {
            // This is fatal.
            ginfra.logPeople.error("Configuration provider not loaded.", kv("error", e.toString()),
                    kv("cause", e.getCause()), kv("url", ginfra.serviceConfig.configUri));
            throw e;
        }

        try {
            ginfra.serviceSpecifics = ServiceConfigSpecifics.LoadServiceConfigSpecifics(ginfra.configProvider,
                    ginfra.serviceConfig.deploymentName, ginfra.serviceConfig.name);
        } catch(Exception e) {
            // This is fatal.
            ginfra.logPeople.error("Service configuration specifics not fetched from configuration provider.",
                    kv("error", e.getMessage()), kv("deployment", ginfra.serviceConfig.deploymentName),
                    kv("service.name", ginfra.serviceConfig.name));
            throw e;
        }
    }

    public static Ginfra getGlobalGinfra() {
        return ginfra;
    }

    public ContainerSpecifics getContainerSpecifics() {
        return containerSpecifics;
    }

    public ServiceConfig getServiceConfig() {
        return serviceConfig;
    }

    public Logger getLogPeople() {
        return logPeople;
    }

    public Logger getLogMachine() {
        return logMachine;
    }

    public Logger getLogAccess() {
        return logAccess;
    }

    public ServiceConfigSpecifics getServiceSpecifics() {
        return serviceSpecifics;
    }

    public Provider getConfigProvider() {
        return configProvider;
    }

}
