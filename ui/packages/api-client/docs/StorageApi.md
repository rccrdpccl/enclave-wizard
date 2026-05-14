# StorageApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfigStorage**](StorageApi.md#getconfigstorage) | **GET** /api/v1/config/storage | Load Block storage configuration |
| [**writeConfigStorage**](StorageApi.md#writeconfigstorage) | **PUT** /api/v1/config/storage | Update Block storage configuration |



## getConfigStorage

> StorageConfig getConfigStorage()

Load Block storage configuration

### Example

```ts
import {
  Configuration,
  StorageApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigStorageRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new StorageApi();

  try {
    const data = await api.getConfigStorage();
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

[**StorageConfig**](StorageConfig.md)

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


## writeConfigStorage

> writeConfigStorage(storageConfig)

Update Block storage configuration

### Example

```ts
import {
  Configuration,
  StorageApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigStorageRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new StorageApi();

  const body = {
    // StorageConfig
    storageConfig: ...,
  } satisfies WriteConfigStorageRequest;

  try {
    const data = await api.writeConfigStorage(body);
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
| **storageConfig** | [StorageConfig](StorageConfig.md) |  | |

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

