# QuayApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfigQuay**](QuayApi.md#getconfigquay) | **GET** /api/v1/config/quay | Load Quay registry configuration |
| [**writeConfigQuay**](QuayApi.md#writeconfigquay) | **PUT** /api/v1/config/quay | Update Quay registry configuration |



## getConfigQuay

> QuayConfig getConfigQuay()

Load Quay registry configuration

### Example

```ts
import {
  Configuration,
  QuayApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigQuayRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new QuayApi();

  try {
    const data = await api.getConfigQuay();
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

[**QuayConfig**](QuayConfig.md)

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


## writeConfigQuay

> writeConfigQuay(quayConfig)

Update Quay registry configuration

### Example

```ts
import {
  Configuration,
  QuayApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigQuayRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new QuayApi();

  const body = {
    // QuayConfig
    quayConfig: ...,
  } satisfies WriteConfigQuayRequest;

  try {
    const data = await api.writeConfigQuay(body);
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
| **quayConfig** | [QuayConfig](QuayConfig.md) |  | |

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

