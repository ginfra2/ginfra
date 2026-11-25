package tech.ginfra.liveconfig;

import com.fasterxml.jackson.annotation.JsonProperty;

public class Meta {
    public int    version;
    public String name;
    public String type;
    public String moduleName;
    public String serviceclass;
    public String invocationcmd;

    @JsonProperty("version")
    public int getVersion() {
        return version;
    }
    public void setVersion(int version) {
        this.version = version;
    }

    @JsonProperty("name")
    public String getName() {
        return name;
    }
    public void setName(String name) {
        this.name = name;
    }

    @JsonProperty("type")
    public String getType() {
      return type;
    }
    public void setType(String type) {
      this.type = type;
    }

    @JsonProperty("module_name")
    public String getModule_name() {
      return moduleName;
    }
    public void setModule_name(String moduleName) {
      this.moduleName = moduleName;
    }

    @JsonProperty("serviceclass")
    public String getServiceclass() {
        return serviceclass;
    }
    public void setServiceclass(String serviceclass) {
        this.serviceclass = serviceclass;
    }

    @JsonProperty("invocationcmd")
    public String getInvocationcmd() {
        return invocationcmd;
    }
    public void setInvocationcmd(String invocationcmd) {
        this.invocationcmd = invocationcmd;
    }

}
