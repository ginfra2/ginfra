package tech.ginfra.api.service_giconfig.provider;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.JsonNodeFactory;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;
import tech.ginfra.api.service_giconfig.client.invoker.ApiResponse;
import tech.ginfra.api.service_giconfig.client.model.*;

import java.util.Base64;


public class APIProvider implements Provider {

    private String auth;
    private DefaultApi api;


    public APIProvider(String url) {
        api = new DefaultApi();
        api.setCustomBaseUrl(url);
    }

    private String path2String(String... path) {
        StringBuilder sb = new StringBuilder();
        for (int idx = 0; idx < path.length; idx++) {
            sb.append(path[idx]);
            if (idx < (path.length-1)) {
                sb.append('.');
            }
        }
        return sb.toString();
    }

    private void checkAuth() throws Exception {
        if (this.auth == null) {
            throw new Exception("AUTH not set for API Configuration Provider.");
        }
    }

    public boolean IsConfigurable() {
        return true;
    }

    public void SetConfigAccessAuth(String auth) {
        this.auth = auth;
    }

    /**
     * Fetches a config from the API based on the given path.
     *
     * @param path The path of the config.
     * @return The fetched config.
     * @throws Exception If fetching the config fails.
     */
    public Config FetchConfig(String... path) throws Exception {
        checkAuth();

        ApiResponse<ConfigQueryResponse> r = api.giConfigQueryConfigWithHttpInfo(path2String(path), auth);
        if (r.getStatusCode() != 200) {
            throw new Exception(String.format("FetchConfig Failed.  Code=%d  Info=%s", r.getStatusCode(), r));
        }
        byte[] decoded = Base64.getDecoder().decode(r.getData().getBlock());

        Config config = new Config();
        ObjectMapper mapper = new ObjectMapper();
        config.config = mapper.readTree(decoded);

        return config;
    }

    public void PostConfig(Config config, String... path) throws Exception {
        checkAuth();

        String j = config.config.toString();
        String b64 = Base64.getEncoder().encodeToString(j.getBytes());
        ConfigSetItem csi = new ConfigSetItem();
        csi.setBlock(b64);
        csi.setPath(path2String(path));

        ConfigServiceSetRequest req = new ConfigServiceSetRequest();
        req.addItemsItem(csi);

        ApiResponse<Void> r = api.giConfigSetConfigWithHttpInfo(auth, req);
        if (r.getStatusCode() != 200) {
            throw new Exception(String.format("PostConfig Failed.  Code=%d  Info=%s", r.getStatusCode(), r));
        }
    }

    public void RemoveConfig(String... path) throws Exception {
        checkAuth();

        ApiResponse<Void> r = api.giRemoveConfigWithHttpInfo(path2String(path), auth);
        if (r.getStatusCode() != 200) {
            throw new Exception(String.format("RemoveConfig Failed.  Code=%d  Info=%s", r.getStatusCode(), r));
        }
    }

    public InitState GetInitState(String deployment, String name, String id) throws Exception {
        checkAuth();

        ApiResponse<ConfigGetInitResponse> r = api.giConfigGetInitWithHttpInfo(auth, deployment, name, id);
        if (r.getStatusCode() != 200) {
            throw new Exception(String.format("GetInitState Failed.  Code=%d  Info=%s", r.getStatusCode(), r));
        }

        return InitState.values()[r.getData().getState()];
    }

    public JsonNode GetState(String deployment, String name) throws Exception {
        checkAuth();

        ApiResponse<ConfigGetStateResponse> r = api.giConfigGetStateWithHttpInfo(auth, deployment, name);
        if (r.getStatusCode() != 200) {
            throw new Exception(String.format("GetState Failed.  Code=%d  Info=%s", r.getStatusCode(), r));
        }

        String b64 = r.getData().getBlock();
        byte[] decoded = Base64.getDecoder().decode(r.getData().getBlock());

        ObjectMapper mapper = new ObjectMapper();
        return mapper.readTree(decoded);
    }

    public void PostState(String deployment, String name, String id, JsonNode state) throws Exception {
        checkAuth();

        String b64 = Base64.getEncoder().encodeToString(state.toString().getBytes());
        ConfigServiceSetStateRequest req = new ConfigServiceSetStateRequest();
        req.setBlock(b64);

        ApiResponse<ConfigSetStateResponse> r = api.giConfigPostStateWithHttpInfo(auth, deployment, name, id, req);
        if (r.getStatusCode() != 200) {
            throw new Exception(String.format("GetState Failed.  Code=%d  Info=%s", r.getStatusCode(), r));
        }
    }

    public String DiscoverService(String deployment, String serviceClass) throws Exception {
        checkAuth();

        ApiResponse<DiscoverServiceResponse> r = api.giDiscoverServiceWithHttpInfo(deployment, serviceClass, auth);
        if (r.getStatusCode() != 200) {
            throw new Exception(String.format("DiscoverService Failed.  Code=%d  Info=%s", r.getStatusCode(), r));
        }

        return r.getData().getUri();
    }

    public String DiscoverPortal(String deployment, String serviceClass) throws Exception {
        checkAuth();

        ApiResponse<DiscoverPortalResponse> r = api.giDiscoverPortalWithHttpInfo(deployment, serviceClass, auth);
        if (r.getStatusCode() != 200) {
            throw new Exception(String.format("DiscoverPortal Failed.  Code=%d  Info=%s", r.getStatusCode(), r));
        }

        return r.getData().getUri();
    }

}
