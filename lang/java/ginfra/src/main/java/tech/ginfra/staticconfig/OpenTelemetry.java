/*
Package tech.ginfra
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Ginfra class
*/
package tech.ginfra.staticconfig;

import tech.ginfra.api.service_giconfig.provider.Config;
import tech.ginfra.api.service_giconfig.provider.Provider;

import java.util.*;

import static tech.ginfra.Ginfra.getGlobalGinfra;

public class OpenTelemetry {

    // --  NAMES -------------------------------------------------------------------------------

    // --  GLOBAL INSTANCE ---------------------------------------------------------------------

    // --  MEMBERS ---------------------------------------------------------------------

    // --  METHODS ---------------------------------------------------------------------

    public static HashMap<String, String> configOpenTelemetry(Provider configProvider, String deployment, String service) {
        var props = new HashMap<String, String>();
        Config disc;
        String uri;

        // Is it configured?
        try {
            disc = Provider.getServiceDiscovery(configProvider, deployment, "otelcol/collector");
            uri = disc.GetValueString("otelcol/collector");
            if (Objects.equals(uri, "")) {
                throw new Exception("Not configured.");
            }

        } catch (Exception e) {
            // Not configured.  Bail.
            props.put("OTEL_SDK_DISABLED", "true");
            return props;
        }

        // We are in business.
        props.put("OTEL_SERVICE_NAME", service);
        props.put("OTEL_TRACES_OTLP_EXPORTER", "otlp");
        props.put("OTEL_METRICS_OTLP_EXPORTER", "otlp");
        props.put("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc");
        props.put("OTEL_EXPORTER_OTLP_ENDPOINT", uri);

        return props;
    }
    
}
