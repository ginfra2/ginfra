package tech.ginfra.liveconfig;

import com.fasterxml.jackson.annotation.JsonProperty;

public class Expose {
    public String name;
    public String serviceclass;

    @JsonProperty("name")
    public String getName() {
        return name;
    }
    public void setName(String name) {
        this.name = name;
    }

    @JsonProperty("serviceclass")
    public String getServiceclass() {
        return serviceclass;
    }
    public void setServiceclass(String serviceclass) {
        this.serviceclass = serviceclass;
    }

}
