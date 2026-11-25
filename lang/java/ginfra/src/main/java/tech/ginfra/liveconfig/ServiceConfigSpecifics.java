package tech.ginfra.liveconfig;

import com.fasterxml.jackson.databind.JsonNode;
import tech.ginfra.api.service_giconfig.provider.Provider;

import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;

public class ServiceConfigSpecifics {

    private int port;
    private Map<String,String> specifics;

    public ServiceConfigSpecifics() {
        specifics = new HashMap<String,String>();
    }

    private static void checkObject(JsonNode node) throws Exception {
        if (!node.isObject()) {
            throw new Exception("Malformed config or specifics.  Service specifics must be a flat JSON object.");
        }
    }

    public static ServiceConfigSpecifics LoadServiceConfigSpecifics(Provider configProvider, String deployment,
        String serviceName) throws Exception {
        ServiceConfigSpecifics spesc = new ServiceConfigSpecifics();

        var cfg = Provider.getServiceConfig(configProvider, deployment, serviceName);
        checkObject(cfg.config);

        spesc.port = Integer.parseInt(cfg.GetValueString("port"));

        var jsonspec = cfg.config.findValue("specifics");
        if (jsonspec == null) {
            throw new Exception("Service specifics missing from service configuration.");
        }
        checkObject(jsonspec);

        for (Iterator<String> it = jsonspec.fieldNames(); it.hasNext(); ) {
            String name = it.next();
            spesc.specifics.put(name, cfg.config.asText(name));
        }

        return spesc;
    }

    public int getPort() {
        return port;
    }

    public Map<String, String> getSpecifics() {
        return specifics;
    }
}
