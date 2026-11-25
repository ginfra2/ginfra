package tech.ginfra.api.service_giconfig.provider;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;
import tech.ginfra.api.service_giconfig.provider.Config;

import static org.junit.jupiter.api.Assertions.*;

public class ConfigTest {
    @Test
    public void testGetConfig() {
        // Arrange
        ObjectMapper objectMapper = new ObjectMapper();
        JsonNode configNode = objectMapper.createObjectNode().put("testName", "testValue");

        Config config = new Config();
        config.config = configNode;

        // Act
        // Assert
        try {
            Config resultConfig = config.GetConfig("testName");
            assertNotNull(resultConfig.config);
            // check if the retrieved config is same as what we put into it
            assertEquals(configNode.findValue("testName"), resultConfig.config);
        } catch (Exception e) {
            fail("Exception should not be thrown");
        }
    }

    @Test
    public void testGetConfigThrowsExceptionIfNoMatch() {
        // Arrange
        ObjectMapper objectMapper = new ObjectMapper();
        JsonNode configNode = objectMapper.createObjectNode().put("testName", "testValue");

        Config config = new Config();
        config.config = configNode;

        // Act
        // Assert
        Exception exception = assertThrows(Exception.class, () -> config.GetConfig("nonExistentName"));
        String expectedMessage = "Config not found.  name:nonExistentName";
        String actualMessage = exception.getMessage();
        assertTrue(actualMessage.contains(expectedMessage));
    }
    @Test
    public void testPutConfigPath() throws Exception {
        // Arrange
        ObjectMapper objectMapper = new ObjectMapper();
        JsonNode configNode = objectMapper.createObjectNode().put("testName", "testValue");
        Config configInitial = new Config();
        configInitial.config = configNode;

        JsonNode newConfig = objectMapper.createObjectNode().put("newConfigName", "newConfigValue");
        Config configToPut = new Config();
        configToPut.config = newConfig;

        // Act
        configInitial.PutConfigPath(configToPut, "testName");

        // Assert
        Config resultConfig = configInitial.GetConfig("testName");
        assertNotNull(resultConfig.config);
        assertEquals(configToPut.config, resultConfig.config);
    }

    @Test
    public void testPutConfigPathEmptyPath() {
        // Arrange
        ObjectMapper objectMapper = new ObjectMapper();
        Config configInitial = new Config();
        configInitial.config = objectMapper.createObjectNode().put("initialName", "initialValue");

        Config configNew = new Config();
        configNew.config = objectMapper.createObjectNode().put("newName", "newValue");

        // Act & Assert
        Exception exception = assertThrows(Exception.class, () -> configInitial.PutConfigPath(configNew));
        String expectedMessage = "Empty path not found.";
        String actualMessage = exception.getMessage();
        assertTrue(actualMessage.contains(expectedMessage));
    }

    @Test
    public void testRemoveConfigPath() throws Exception {
        // Arrange
        ObjectMapper objectMapper = new ObjectMapper();
        JsonNode configNode = objectMapper.createObjectNode().put("testName", "testValue");
        Config config = new Config();
        config.config = configNode;

        // Act
        Config removedConfig = config.RemoveConfigPath("testName");

        // Assert
        assertNotNull(removedConfig.config);
        assertEquals("testValue", removedConfig.config.asText());
    }


}
