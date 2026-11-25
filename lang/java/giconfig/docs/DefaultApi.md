# DefaultApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**giConfigDeregisterService**](DefaultApi.md#giConfigDeregisterService) | **POST** /configservice/deregister/{deployment}/{name}/ | Deregister a service. |
| [**giConfigDiscoveryAdd**](DefaultApi.md#giConfigDiscoveryAdd) | **PUT** /configservice/discovery/add/ | Add a discovery point |
| [**giConfigGetInit**](DefaultApi.md#giConfigGetInit) | **GET** /configservice/state/init/{deployment}/{name}/{id}/ | Get the current initializaton state. |
| [**giConfigGetState**](DefaultApi.md#giConfigGetState) | **GET** /configservice/state/get/{deployment}/{name}/ | Get the current service state. |
| [**giConfigPortal**](DefaultApi.md#giConfigPortal) | **POST** /configservice/portal/{operation}/{deployment}/ | Register or update a portal. |
| [**giConfigPostState**](DefaultApi.md#giConfigPostState) | **POST** /configservice/state/post/{deployment}/{name}/{id} | Set State. |
| [**giConfigQueryConfig**](DefaultApi.md#giConfigQueryConfig) | **GET** /configservice/query | Query config. |
| [**giConfigRegisterInstance**](DefaultApi.md#giConfigRegisterInstance) | **POST** /configservice/instance | Register or update a service instance. |
| [**giConfigRegisterService**](DefaultApi.md#giConfigRegisterService) | **POST** /configservice/register | Register a service. |
| [**giConfigSetConfig**](DefaultApi.md#giConfigSetConfig) | **POST** /configservice/set | Set configurations. |
| [**giConfigStoreGet**](DefaultApi.md#giConfigStoreGet) | **POST** /configservice/store/get | Get a file. |
| [**giConfigStorePut**](DefaultApi.md#giConfigStorePut) | **POST** /configservice/store/put | Store a file. |
| [**giDiscoverPortal**](DefaultApi.md#giDiscoverPortal) | **GET** /configservice/discover/portal/{deployment}/{serviceclass} | Discover a portal for a deployment. |
| [**giDiscoverService**](DefaultApi.md#giDiscoverService) | **GET** /configservice/discover/service/{deployment}/{serviceclass} | Discover a service. |
| [**giRemoveConfig**](DefaultApi.md#giRemoveConfig) | **GET** /configservice/remove | Remove config. |


<a id="giConfigDeregisterService"></a>
# **giConfigDeregisterService**
> ConfigServiceRegistrationResponse giConfigDeregisterService(confauth, name, deployment)

Deregister a service.

A priviledged deregistration request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Administrative access token given to the config service when started.
    String name = "name_example"; // String | Unique name of the service.
    String deployment = "deployment_example"; // String | Unique deployment name.
    try {
      ConfigServiceRegistrationResponse result = apiInstance.giConfigDeregisterService(confauth, name, deployment);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigDeregisterService");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Administrative access token given to the config service when started. | |
| **name** | **String**| Unique name of the service. | |
| **deployment** | **String**| Unique deployment name. | |

### Return type

[**ConfigServiceRegistrationResponse**](ConfigServiceRegistrationResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigDiscoveryAdd"></a>
# **giConfigDiscoveryAdd**
> giConfigDiscoveryAdd(confauth, deployment, sclass, uri)

Add a discovery point

It will add a discovery point.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Administrative access token given to the config service when started.
    String deployment = "deployment_example"; // String | Unique deployment name.
    String sclass = "sclass_example"; // String | Service class to add.
    String uri = "uri_example"; // String | Uri for the service class.
    try {
      apiInstance.giConfigDiscoveryAdd(confauth, deployment, sclass, uri);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigDiscoveryAdd");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Administrative access token given to the config service when started. | |
| **deployment** | **String**| Unique deployment name. | |
| **sclass** | **String**| Service class to add. | |
| **uri** | **String**| Uri for the service class. | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigGetInit"></a>
# **giConfigGetInit**
> ConfigGetInitResponse giConfigGetInit(confauth, deployment, name, id)

Get the current initializaton state.

An initializaton state request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Service access token.
    String deployment = "deployment_example"; // String | Unique deployment name.
    String name = "name_example"; // String | Unique name of the service.
    String id = "id_example"; // String | Unique identity of the calling instance.
    try {
      ConfigGetInitResponse result = apiInstance.giConfigGetInit(confauth, deployment, name, id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigGetInit");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Service access token. | |
| **deployment** | **String**| Unique deployment name. | |
| **name** | **String**| Unique name of the service. | |
| **id** | **String**| Unique identity of the calling instance. | |

### Return type

[**ConfigGetInitResponse**](ConfigGetInitResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigGetState"></a>
# **giConfigGetState**
> ConfigGetStateResponse giConfigGetState(confauth, deployment, name)

Get the current service state.

An service state request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Service access token.
    String deployment = "deployment_example"; // String | Unique deployment name.
    String name = "name_example"; // String | Unique name of the service.
    try {
      ConfigGetStateResponse result = apiInstance.giConfigGetState(confauth, deployment, name);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigGetState");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Service access token. | |
| **deployment** | **String**| Unique deployment name. | |
| **name** | **String**| Unique name of the service. | |

### Return type

[**ConfigGetStateResponse**](ConfigGetStateResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigPortal"></a>
# **giConfigPortal**
> giConfigPortal(confauth, operation, deployment, configPortalRequest)

Register or update a portal.

Register or update a portal.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Administrative access token given to the config service when started.
    String operation = "put"; // String | Unique deployment name.
    String deployment = "deployment_example"; // String | Unique deployment name.
    ConfigPortalRequest configPortalRequest = new ConfigPortalRequest(); // ConfigPortalRequest | Register a portal.
    try {
      apiInstance.giConfigPortal(confauth, operation, deployment, configPortalRequest);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigPortal");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Administrative access token given to the config service when started. | |
| **operation** | **String**| Unique deployment name. | [enum: put, remove] |
| **deployment** | **String**| Unique deployment name. | |
| **configPortalRequest** | [**ConfigPortalRequest**](ConfigPortalRequest.md)| Register a portal. | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigPostState"></a>
# **giConfigPostState**
> ConfigSetStateResponse giConfigPostState(confauth, deployment, name, id, configServiceSetStateRequest)

Set State.

A set state request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Access token.
    String deployment = "deployment_example"; // String | Unique deployment name.
    String name = "name_example"; // String | Unique name of the service.
    String id = "id_example"; // String | Unique identity of the calling instance.
    ConfigServiceSetStateRequest configServiceSetStateRequest = new ConfigServiceSetStateRequest(); // ConfigServiceSetStateRequest | Set config.
    try {
      ConfigSetStateResponse result = apiInstance.giConfigPostState(confauth, deployment, name, id, configServiceSetStateRequest);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigPostState");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Access token. | |
| **deployment** | **String**| Unique deployment name. | |
| **name** | **String**| Unique name of the service. | |
| **id** | **String**| Unique identity of the calling instance. | |
| **configServiceSetStateRequest** | [**ConfigServiceSetStateRequest**](ConfigServiceSetStateRequest.md)| Set config. | |

### Return type

[**ConfigSetStateResponse**](ConfigSetStateResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigQueryConfig"></a>
# **giConfigQueryConfig**
> ConfigQueryResponse giConfigQueryConfig(qpath, confauth)

Query config.

Query configuration data.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String qpath = "qpath_example"; // String | The dot separated path the configuration point.
    String confauth = "confauth_example"; // String | Access token.
    try {
      ConfigQueryResponse result = apiInstance.giConfigQueryConfig(qpath, confauth);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigQueryConfig");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **qpath** | **String**| The dot separated path the configuration point. | |
| **confauth** | **String**| Access token. | |

### Return type

[**ConfigQueryResponse**](ConfigQueryResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **400** | Bad request |  -  |
| **404** | Not found |  -  |
| **500** | InternalError |  -  |

<a id="giConfigRegisterInstance"></a>
# **giConfigRegisterInstance**
> ConfigServiceRegistrationResponse giConfigRegisterInstance(confauth, configServiceInstanceRequest)

Register or update a service instance.

A privileged registration request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Administrative access token given to the config service when started.
    ConfigServiceInstanceRequest configServiceInstanceRequest = new ConfigServiceInstanceRequest(); // ConfigServiceInstanceRequest | Register a service instance.
    try {
      ConfigServiceRegistrationResponse result = apiInstance.giConfigRegisterInstance(confauth, configServiceInstanceRequest);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigRegisterInstance");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Administrative access token given to the config service when started. | |
| **configServiceInstanceRequest** | [**ConfigServiceInstanceRequest**](ConfigServiceInstanceRequest.md)| Register a service instance. | |

### Return type

[**ConfigServiceRegistrationResponse**](ConfigServiceRegistrationResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigRegisterService"></a>
# **giConfigRegisterService**
> ConfigServiceRegistrationResponse giConfigRegisterService(confauth, configServiceRegistrationRequest)

Register a service.

A privileged registration request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Administrative access token given to the config service when started.
    ConfigServiceRegistrationRequest configServiceRegistrationRequest = new ConfigServiceRegistrationRequest(); // ConfigServiceRegistrationRequest | Register a service.
    try {
      ConfigServiceRegistrationResponse result = apiInstance.giConfigRegisterService(confauth, configServiceRegistrationRequest);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigRegisterService");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Administrative access token given to the config service when started. | |
| **configServiceRegistrationRequest** | [**ConfigServiceRegistrationRequest**](ConfigServiceRegistrationRequest.md)| Register a service. | |

### Return type

[**ConfigServiceRegistrationResponse**](ConfigServiceRegistrationResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigSetConfig"></a>
# **giConfigSetConfig**
> giConfigSetConfig(confauth, configServiceSetRequest)

Set configurations.

A privileged registration request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Access token.
    ConfigServiceSetRequest configServiceSetRequest = new ConfigServiceSetRequest(); // ConfigServiceSetRequest | Set config.
    try {
      apiInstance.giConfigSetConfig(confauth, configServiceSetRequest);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigSetConfig");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Access token. | |
| **configServiceSetRequest** | [**ConfigServiceSetRequest**](ConfigServiceSetRequest.md)| Set config. | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigStoreGet"></a>
# **giConfigStoreGet**
> File giConfigStoreGet(confauth, name, raw, deployment, service)

Get a file.

A get a file request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Access token.
    String name = "name_example"; // String | File store name.
    Boolean raw = true; // Boolean | Raw or processed.
    String deployment = "deployment_example"; // String | Unique deployment name or empty
    String service = "service_example"; // String | Unique name of the service.  It may be empty.
    try {
      File result = apiInstance.giConfigStoreGet(confauth, name, raw, deployment, service);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigStoreGet");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Access token. | |
| **name** | **String**| File store name. | |
| **raw** | **Boolean**| Raw or processed. | |
| **deployment** | **String**| Unique deployment name or empty | [optional] |
| **service** | **String**| Unique name of the service.  It may be empty. | [optional] |

### Return type

[**File**](File.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giConfigStorePut"></a>
# **giConfigStorePut**
> giConfigStorePut(confauth, name, body, deployment, service)

Store a file.

A store a file request.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String confauth = "confauth_example"; // String | Access token.
    String name = "name_example"; // String | File store name.
    File body = new File("/path/to/file"); // File | Data body.
    String deployment = "deployment_example"; // String | Unique deployment name or empty
    String service = "service_example"; // String | Service name or empty
    try {
      apiInstance.giConfigStorePut(confauth, name, body, deployment, service);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giConfigStorePut");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **confauth** | **String**| Access token. | |
| **name** | **String**| File store name. | |
| **body** | **File**| Data body. | |
| **deployment** | **String**| Unique deployment name or empty | [optional] |
| **service** | **String**| Service name or empty | [optional] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/octet-stream
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **403** | Not authorized |  -  |
| **400** | Bad request |  -  |
| **500** | InternalError |  -  |

<a id="giDiscoverPortal"></a>
# **giDiscoverPortal**
> DiscoverPortalResponse giDiscoverPortal(deployment, serviceclass, confauth)

Discover a portal for a deployment.

Discover a portal location for a deployment.  A portal provides external access to a service.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String deployment = "deployment_example"; // String | Deployment name.
    String serviceclass = "serviceclass_example"; // String | Service class name.
    String confauth = "confauth_example"; // String | Access token.
    try {
      DiscoverPortalResponse result = apiInstance.giDiscoverPortal(deployment, serviceclass, confauth);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giDiscoverPortal");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **deployment** | **String**| Deployment name. | |
| **serviceclass** | **String**| Service class name. | |
| **confauth** | **String**| Access token. | |

### Return type

[**DiscoverPortalResponse**](DiscoverPortalResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **400** | Bad request |  -  |
| **404** | Not found |  -  |
| **500** | InternalError |  -  |

<a id="giDiscoverService"></a>
# **giDiscoverService**
> DiscoverServiceResponse giDiscoverService(deployment, serviceclass, confauth)

Discover a service.

Discover a service entrypoint uri.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String deployment = "deployment_example"; // String | Deployment name.
    String serviceclass = "serviceclass_example"; // String | Service class name.
    String confauth = "confauth_example"; // String | Access token.
    try {
      DiscoverServiceResponse result = apiInstance.giDiscoverService(deployment, serviceclass, confauth);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giDiscoverService");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **deployment** | **String**| Deployment name. | |
| **serviceclass** | **String**| Service class name. | |
| **confauth** | **String**| Access token. | |

### Return type

[**DiscoverServiceResponse**](DiscoverServiceResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **400** | Bad request |  -  |
| **404** | Not found |  -  |
| **500** | InternalError |  -  |

<a id="giRemoveConfig"></a>
# **giRemoveConfig**
> giRemoveConfig(qpath, confauth)

Remove config.

Remove configuration data.

### Example
```java
// Import classes:
import tech.ginfra.api.service_giconfig.client.invoker.ApiClient;
import tech.ginfra.api.service_giconfig.client.invoker.ApiException;
import tech.ginfra.api.service_giconfig.client.invoker.Configuration;
import tech.ginfra.api.service_giconfig.client.invoker.models.*;
import tech.ginfra.api.service_giconfig.client.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String qpath = "qpath_example"; // String | The dot separated path the configuration point.
    String confauth = "confauth_example"; // String | Access token.
    try {
      apiInstance.giRemoveConfig(qpath, confauth);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#giRemoveConfig");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **qpath** | **String**| The dot separated path the configuration point. | |
| **confauth** | **String**| Access token. | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Success |  -  |
| **400** | Bad request |  -  |
| **404** | Not found |  -  |
| **500** | InternalError |  -  |

