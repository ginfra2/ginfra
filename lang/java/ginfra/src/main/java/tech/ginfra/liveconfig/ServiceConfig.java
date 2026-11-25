package tech.ginfra.liveconfig;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.nio.file.Paths;
import java.util.Objects;

public class ServiceConfig {

    // Config
    public String version;

    // Values
    public InstanceType instanceType;
    public String auth;
    public String name;
    public String deploymentName;
    public String configUri;
    public String configPortalUri;
    public String configToken;
    public String id;
    public int port;
    public boolean clear;
    public JsonNode specifics;

    // Loaded
    public String instanceTypeLoaded;
    public String clearLoaded;

    // Not loaded by this class.
    public String serviceHome;

    @JsonProperty("version")
    public String getVersion() {
        return version;
    }
    public void setVersion(String ver) {
        this.version = ver;
    }

    @JsonProperty("auth")
    public String getAuth() {
        return auth;
    }
    public void setAuth(String auth) {
        this.auth = auth;
    }

    @JsonProperty("name")
    public String getName() {
        return name;
    }
    public void setName(String name) {
        this.name = name;
    }

    @JsonProperty("deployment")
    public String getDeploymentName() {
        return deploymentName;
    }
    public void setDeploymentName(String deploymentName) {
        this.deploymentName = deploymentName;
    }

    @JsonProperty("config_uri")
    public String getConfigUri() {
        return configUri;
    }
    public void setConfigUri(String configUri) {
        this.configUri = configUri;
    }

    @JsonProperty("config_portal_uri")
    public String getConfigPortalUri() {
        return configPortalUri;
    }
    public void setConfigPortalUri(String configPortalUri) {
        this.configPortalUri = configPortalUri;
    }

    @JsonProperty("config_access_token")
    public String getConfigToken() {
        return configToken;
    }
    public void setConfigToken(String configToken) {
        this.configToken = configToken;
    }

    @JsonProperty("id")
    public String getId() {
        return id;
    }
    public void setId(String id) {
        this.id = id;
    }

    @JsonProperty("type")
    public String getInstanceTypeLoaded() {
        return instanceTypeLoaded;
    }
    public void setInstanceTypeLoaded(String instanceTypeLoaded) {
        this.instanceTypeLoaded = instanceTypeLoaded;
    }

    @JsonProperty("port")
    public int getPort() {
        return port;
    }
    public void setPort(int port) {
        this.port = port;
    }

    @JsonProperty("clear")
    public String getClearLoaded() {
        return clearLoaded;
    }
    public void setClearLoaded(String clearLoaded) {
        this.clearLoaded = clearLoaded;
    }

    @JsonProperty("specifics")
    public JsonNode getSpecifics() {
        return specifics;
    }
    public void setSpecifics(JsonNode specifics) {
        this.specifics = specifics;
    }

    public String getServiceHome() {
        return serviceHome;
    }

    public void setServiceHome(String serviceHome) {
        this.serviceHome = serviceHome;
    }

    public static ServiceConfig LoadServiceConfig(String path) throws Exception {
        var mapper = new ObjectMapper();
        var serviceConfig = mapper.readValue(Paths.get(path).toFile(), ServiceConfig.class);

        if ((serviceConfig.getClearLoaded() != null)&&(!Objects.equals(serviceConfig.getClearLoaded(), ""))) {
            serviceConfig.clear = true;
        }

        if ((serviceConfig.getInstanceTypeLoaded() != null)&&(!Objects.equals(serviceConfig.getInstanceTypeLoaded(), ""))) {
            switch(serviceConfig.getInstanceTypeLoaded().toLowerCase()) {
                case "local":
                    serviceConfig.instanceType = InstanceType.ITLocal;
                    break;
                case "host":
                    serviceConfig.instanceType = InstanceType.ITHost;
                    break;
                case "managed":
                    serviceConfig.instanceType = InstanceType.ITManaged;
                    break;
                default:
                    serviceConfig.instanceType = InstanceType.ITBad;
                    break;
            }
        }

        return serviceConfig;
    }

}
