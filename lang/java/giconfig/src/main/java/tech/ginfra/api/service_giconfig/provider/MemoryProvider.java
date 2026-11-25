package tech.ginfra.api.service_giconfig.provider;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.JsonNodeFactory;

import java.util.Objects;

/**
 * A class that implements the Provider interface and provides in-memory storage for configurations and states.
 * This implementation is not threadsafe.
 */
public class MemoryProvider implements Provider {

    protected Config config;

    public MemoryProvider() {
        config = new Config();
    }

    public boolean IsConfigurable() {
        return true;
    }

    public void SetConfigAccessAuth(String auth) {
        // NOP
    }

    public Config FetchConfig(String... path) throws Exception {
        return config.GetConfigPath(path);
    }

    public void PostConfig(Config config, String... path) throws Exception {
        config.PutConfigPath(config, path);
    }

    public void RemoveConfig(String... path) throws Exception {
        config.RemoveConfigPath(path);
    }

    public InitState GetInitState(String deployment, String name, String id) throws Exception {
        // For standalone testing, so it is always initializing.
        return InitState.InitNew;
    }

    public JsonNode GetState(String deployment, String name) throws Exception {
        return JsonNodeFactory.instance.objectNode();
    }

    public void PostState(String deployment, String name, String id, JsonNode state) throws Exception {
        // NOP
    }

    public String DiscoverService(String deployment, String serviceClass) throws Exception {
        Config config = Provider.getServiceDiscovery(this, deployment, serviceClass);
        return config.GetValueString(serviceClass);
    }

    public String DiscoverPortal(String deployment, String serviceClass) throws Exception {
        Config config = Provider.getPortalDiscovery(this, deployment, serviceClass);
        String value = config.GetValueString(serviceClass);
        if (Objects.equals(value, "")) {
            value = config.GetValueString("*");
        }
        return value;
    }

}
