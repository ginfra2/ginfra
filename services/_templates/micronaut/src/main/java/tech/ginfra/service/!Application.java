package tech.ginfra.service;

import io.micronaut.context.env.PropertySource;
import io.micronaut.runtime.Micronaut;
import tech.ginfra.Ginfra;

import java.util.HashMap;
import java.util.Map;

public class Application {

    public static void main(String[] args) {

        if (args.length < 1) {
            System.out.println("Not enough arguments.");
            System.exit(1);
        }

        try {
            Ginfra.LoadGinfra(args[0]);
        } catch (Exception e) {
            System.out.println("Failed to start service.");
            System.exit(1);
        }

        // Run
        var mn = Micronaut.build(args);
        mn.classes(Application.class);

        Map<String, Object> props = new HashMap<>();
        props.put("micronaut.server.port", Ginfra.getGlobalGinfra().getServiceSpecifics().getPort());

        var ps = PropertySource.of("MySource", props);
        mn.propertySources(ps);
        mn.start();
    }
}