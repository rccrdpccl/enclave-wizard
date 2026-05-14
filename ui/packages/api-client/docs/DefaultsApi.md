# DefaultsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getDefaults**](DefaultsApi.md#getdefaults) | **GET** /api/v1/defaults | Get default configuration values |



## getDefaults

> Defaults getDefaults()

Get default configuration values

Returns default values read from defaults/deployment.yaml and plugin definitions.

### Example

```ts
import {
  Configuration,
  DefaultsApi,
} from '@enclave-wizard-ui/api-client';
import type { GetDefaultsRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new DefaultsApi();

  try {
    const data = await api.getDefaults();
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

[**Defaults**](Defaults.md)

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

