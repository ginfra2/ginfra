# SERVICES

A service is anything deployed by ginfra.  

## Deploying Services

## Types of Services

| Deployment Type | Description                                                                                                                                      |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| Local           | Local deployments are started on the local machine as a local process.                                                                           |
| Host            | Container deployments on a single docker engine host.                                                                                            |
| Managed         | Bare-metal kubernetes is directly supported in open source Ginfra, but others have successfully implemented EKS (unfortunately not open source). |

## Service lifestyle

1) Start command initiated by ginfra.

2) Ginfra places a service_config.json in the service's root directory.

This file has the bare minimum to start the service.  Technically, the service doesn't need to use this at all,
but most of the provided services and generated services do.

```{
   "auth": "XXXXXXXXXXXX",
   "clear": "false",
   "config_access_token": "AAAAAAAAAAAA",
   "config_host_uri": "config://",
   "config_uri": "config://",
   "deployment": "default",
   "name": "server_otelcol",
   "port": 8900,
   "type": "local"
}
```

|         Key         | What it does                                                                                                                  |
|:-------------------:|-------------------------------------------------------------------------------------------------------------------------------|
|        auth         | This is the authorization key for allowing service calls to control the service.                                              |
|        clear        | Notifies service to clear its stores and caches.  This is mostly for testing and should be used and implemented with caution. |
| config_access_token | The access token for accessing the configuration server                                                                       |
|   config_host_uri   | For connecting to the configuration service.  The ginfra configuration providers know how to use it.                          |
|     config_uri      | Alternate local uri for the config service.  This is mostly for debugging and should not be used.                             |
|     deployment      | Deployment name for the service.                                                                                              |
|        name         | Service name for the service.                                                                                                 |
|        port         | Port for the initial service.  At a minimum, this should be the port exposed by a controlling interface.                      |
|        type         | [Type of instance](## Types of instances).                                                                                    |

3) Ginfra tries to find a [specific.json](Specifics.md) file in to get hints about how to start the service.

This is not really useful for local processes.  

4) The service process or container is started


