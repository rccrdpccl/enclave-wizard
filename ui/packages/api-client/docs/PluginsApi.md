# PluginsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfigPlugins**](PluginsApi.md#getconfigplugins) | **GET** /api/v1/config/plugins | Load Enabled plugins configuration |
| [**listPlugins**](PluginsApi.md#listplugins) | **GET** /api/v1/plugins | List available plugins |
| [**validatePluginCombination**](PluginsApi.md#validateplugincombination) | **POST** /api/v1/plugins/validate | Validate plugin combination |
| [**writeConfigPlugins**](PluginsApi.md#writeconfigplugins) | **PUT** /api/v1/config/plugins | Update Enabled plugins configuration |



## getConfigPlugins

> PluginsConfig getConfigPlugins()

Load Enabled plugins configuration

### Example

```ts
import {
  Configuration,
  PluginsApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigPluginsRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new PluginsApi();

  try {
    const data = await api.getConfigPlugins();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**PluginsConfig**](PluginsConfig.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listPlugins

> PluginsOutputBody listPlugins()

List available plugins

Returns all known plugins and their types.

### Example

```ts
import {
  Configuration,
  PluginsApi,
} from '@enclave-wizard-ui/api-client';
import type { ListPluginsRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new PluginsApi();

  try {
    const data = await api.listPlugins();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**PluginsOutputBody**](PluginsOutputBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## validatePluginCombination

> PluginValidateOutputBody validatePluginCombination(pluginValidateInputBody)

Validate plugin combination

Checks whether the given set of plugins forms a valid deployment combination.

### Example

```ts
import {
  Configuration,
  PluginsApi,
} from '@enclave-wizard-ui/api-client';
import type { ValidatePluginCombinationRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new PluginsApi();

  const body = {
    // PluginValidateInputBody
    pluginValidateInputBody: ...,
  } satisfies ValidatePluginCombinationRequest;

  try {
    const data = await api.validatePluginCombination(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **pluginValidateInputBody** | [PluginValidateInputBody](PluginValidateInputBody.md) |  | |

### Return type

[**PluginValidateOutputBody**](PluginValidateOutputBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## writeConfigPlugins

> writeConfigPlugins(pluginsConfig)

Update Enabled plugins configuration

### Example

```ts
import {
  Configuration,
  PluginsApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigPluginsRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new PluginsApi();

  const body = {
    // PluginsConfig
    pluginsConfig: ...,
  } satisfies WriteConfigPluginsRequest;

  try {
    const data = await api.writeConfigPlugins(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **pluginsConfig** | [PluginsConfig](PluginsConfig.md) |  | |

### Return type

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

