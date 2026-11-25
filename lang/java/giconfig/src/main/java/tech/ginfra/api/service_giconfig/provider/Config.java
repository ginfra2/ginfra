package tech.ginfra.api.service_giconfig.provider;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

public class Config {
    public JsonNode config;

    public Config GetConfig(String name) throws Exception {
        Config result = new Config();

        result.config = config.findValue(name);
        if (result.config == null) {
            throw new Exception("Config not found.  name:" + name);
        }
        return result;
    }

    public Config GetConfigPath(String... path) throws Exception {
        Config result = new Config();

        if (path.length < 1) {
            throw new Exception("Empty path not found.");
        }

        result.config = config;
        for (String nodename : path)
        {
            result = result.GetConfig(nodename);
        }
        return result;
    }

    public void PutConfigPath(Config inconfig, String... path) throws Exception {
        if (path.length < 1) {
            throw new Exception("Empty path not found.");
        }
        JsonNode current = config;
        for (int i = 0 ; i < (path.length-1); i++) {
            current = current.findPath(path[i]);
        }
        ((ObjectNode)current).set(path[path.length-1], inconfig.config);
    }

    public Config RemoveConfigPath(String... path) throws Exception {
        if (path.length < 1) {
            throw new Exception("Empty path not found.");
        }
        JsonNode current = config;
        for (int i = 0 ; i < (path.length-1); i++) {
            current = current.findPath(path[i]);
        }

        Config result = new Config();
        result.config = ((ObjectNode)current).remove(path[path.length-1]);
        return result;
    }

    public void PutValue(String name, String value) throws Exception {
        ((ObjectNode)config).put(name, value);
    }

    public void PutValue(String name, int value) throws Exception {
        ((ObjectNode)config).put(name, value);
    }

    public void PutValue(String name, boolean value) throws Exception {
        ((ObjectNode)config).put(name, value);
    }

    public void PutValue(String name, float value) throws Exception {
        ((ObjectNode)config).put(name, value);
    }

    public String GetValueString(String name) {
        JsonNode item = config.findValue(name);
        if (item != null) {
            return item.asText();
        }
        return "";
    }

    public long GetValueLong(String name) {
        return ((ObjectNode)config).asLong();
    }
}
