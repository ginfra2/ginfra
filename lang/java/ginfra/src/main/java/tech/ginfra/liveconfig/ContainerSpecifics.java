package tech.ginfra.liveconfig;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.nio.file.Paths;
import java.nio.file.Files;
import java.util.HashMap;

import static tech.ginfra.Ginfra.SPECIFICS_FILE_NAME;

public class ContainerSpecifics {

    @JsonProperty("exposed")
    public HashMap<String, Expose> exposed;

    @JsonProperty("meta")
    public Meta meta;

    public String raw;

    public ContainerSpecifics() {
        exposed = new HashMap<String, Expose>();
    }

    public static ContainerSpecifics LoadSpecifics(String path) throws Exception {
        var p = Paths.get(path);
        if (!Files.exists(p)) {
            // Try the root dir.
            p = Paths.get("/", SPECIFICS_FILE_NAME);
            if (!Files.exists(p)) {
                throw new Exception("Specifics file does not exist.");
            }
        }

        var content = new String(Files.readAllBytes(p));

        var mapper = new ObjectMapper();
        var cs = mapper.readValue(content, ContainerSpecifics.class);
        cs.raw = content;

        return cs;
    }

}
