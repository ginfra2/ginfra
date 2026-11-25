package tech.ginfra.api.service_giconfig.provider;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.JsonNodeFactory;

import java.net.URL;
import java.nio.file.Paths;

/**
 * A class that implements the Provider interface and provides in-memory storage for configurations and states
 * but is initially loaded from a file.
 * This implementation is not threadsafe.
 */
public class FileProvider extends MemoryProvider {

    public FileProvider(String url) throws Exception {
        super();

        URL purl = new URL(url);
        ObjectMapper mapper = new ObjectMapper();
        config.config = mapper.readTree(Paths.get(purl.getHost(), purl.getPath()).toFile());

    }

}
