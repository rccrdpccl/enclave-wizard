# NetworkApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfigNetwork**](NetworkApi.md#getconfignetwork) | **GET** /api/v1/config/network | Load Host network configuration |
| [**writeConfigNetwork**](NetworkApi.md#writeconfignetwork) | **PUT** /api/v1/config/network | Update Host network configuration |



## getConfigNetwork

> NetworkConfig getConfigNetwork()

Load Host network configuration

### Example

```ts
import {
  Configuration,
  NetworkApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigNetworkRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new NetworkApi();

  try {
    const data = await api.getConfigNetwork();
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

[**NetworkConfig**](NetworkConfig.md)

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


## writeConfigNetwork

> writeConfigNetwork(networkConfig)

Update Host network configuration

### Example

```ts
import {
  Configuration,
  NetworkApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigNetworkRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new NetworkApi();

  const body = {
    // NetworkConfig
    networkConfig: ...,
  } satisfies WriteConfigNetworkRequest;

  try {
    const data = await api.writeConfigNetwork(body);
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
| **networkConfig** | [NetworkConfig](NetworkConfig.md) |  | |

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

