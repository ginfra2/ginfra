package tech.ginfra.api.service_giconfig.provider;

import com.fasterxml.jackson.databind.JsonNode;

public interface Provider {
    boolean IsConfigurable();
    void SetConfigAccessAuth(String auth);
    Config FetchConfig(String... path) throws Exception;
    void PostConfig(Config config, String... path) throws Exception;
    void RemoveConfig(String... path) throws Exception;
    InitState GetInitState(String deployment, String name, String id) throws Exception;
    JsonNode GetState(String deployment, String name) throws Exception;
    void PostState(String deployment, String name, String id, JsonNode state) throws Exception;
    String DiscoverService(String deployment, String serviceClass) throws Exception;
    String DiscoverPortal(String deployment, String serviceClass) throws Exception;

    static Provider getProvider(String url) throws Exception {
        Provider r;

        try {
            if ((url.startsWith("http://")) || (url.startsWith("https://"))) {
                r = new APIProvider(url);
            } else if (url.startsWith("file://")) {
                r = new FileProvider(url);
            } else if (url.startsWith("memory://")) {
                r = new MemoryProvider();
            } else {
                throw new Exception("Unsupported provider type: " + url);
            }
        } catch (Exception e) {
            throw new Exception("Failed to load configuration provider", e);
        }
        return r;
    }

    static Config getServiceConfig(Provider provider, String deployment, String serviceName) throws Exception {
        return provider.FetchConfig("deployments", deployment, "services", serviceName, "service", "config");
    }

    static Config getServiceDiscovery(Provider provider, String deployment, String serviceClass) throws Exception {
        return provider.FetchConfig("deployments", deployment, "discovery", "services", serviceClass);
    }

    static Config getPortalDiscovery(Provider provider, String deployment, String serviceClass) throws Exception {
        Config config;
        try {
            config = provider.FetchConfig("deployments", deployment, "discovery", "portals", serviceClass);
        } catch (Exception e) {
            // Try catch-all
            try {
                config = provider.FetchConfig("deployments", deployment, "discovery", "portals", "*");
            } catch (Exception ee) {
                throw e;
            }
        }
        return config;
    }
}
