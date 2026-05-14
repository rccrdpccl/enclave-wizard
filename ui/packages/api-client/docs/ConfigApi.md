# ConfigApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfig**](ConfigApi.md#getconfig) | **GET** /api/v1/config | Load existing configuration |
| [**previewConfig**](ConfigApi.md#previewconfig) | **POST** /api/v1/config/preview | Preview rendered YAML |
| [**validateConfig**](ConfigApi.md#validateconfig) | **POST** /api/v1/config/validate | Validate configuration |
| [**writeConfig**](ConfigApi.md#writeconfig) | **PUT** /api/v1/config | Write configuration to disk |



## getConfig

> EnclaveConfig getConfig()

Load existing configuration

Reads config/global.yaml, config/certificates.yaml, and config/cloud_infra.yaml from the Enclave directory and returns the merged configuration.

### Example

```ts
import {
  Configuration,
  ConfigApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new ConfigApi();

  try {
    const data = await api.getConfig();
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

[**EnclaveConfig**](EnclaveConfig.md)

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


## previewConfig

> PreviewConfigOutputBody previewConfig(enclaveConfig)

Preview rendered YAML

Returns the rendered YAML content for each config file without writing to disk.

### Example

```ts
import {
  Configuration,
  ConfigApi,
} from '@enclave-wizard-ui/api-client';
import type { PreviewConfigRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new ConfigApi();

  const body = {
    // EnclaveConfig
    enclaveConfig: ...,
  } satisfies PreviewConfigRequest;

  try {
    const data = await api.previewConfig(body);
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
| **enclaveConfig** | [EnclaveConfig](EnclaveConfig.md) |  | |

### Return type

[**PreviewConfigOutputBody**](PreviewConfigOutputBody.md)

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


## validateConfig

> ValidateConfigOutputBody validateConfig(enclaveConfig)

Validate configuration

Validates the candidate configuration against Enclave JSON schemas and returns structured errors.

### Example

```ts
import {
  Configuration,
  ConfigApi,
} from '@enclave-wizard-ui/api-client';
import type { ValidateConfigRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new ConfigApi();

  const body = {
    // EnclaveConfig
    enclaveConfig: ...,
  } satisfies ValidateConfigRequest;

  try {
    const data = await api.validateConfig(body);
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
| **enclaveConfig** | [EnclaveConfig](EnclaveConfig.md) |  | |

### Return type

[**ValidateConfigOutputBody**](ValidateConfigOutputBody.md)

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


## writeConfig

> writeConfig(enclaveConfig)

Write configuration to disk

Accepts wizard state, serializes to YAML, and writes to the Enclave config directory.

### Example

```ts
import {
  Configuration,
  ConfigApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new ConfigApi();

  const body = {
    // EnclaveConfig
    enclaveConfig: ...,
  } satisfies WriteConfigRequest;

  try {
    const data = await api.writeConfig(body);
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
| **enclaveConfig** | [EnclaveConfig](EnclaveConfig.md) |  | |

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

