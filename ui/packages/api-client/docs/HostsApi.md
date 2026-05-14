# HostsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfigHosts**](HostsApi.md#getconfighosts) | **GET** /api/v1/config/hosts | Load Discovery hosts (cloud infrastructure) |
| [**writeConfigHosts**](HostsApi.md#writeconfighosts) | **PUT** /api/v1/config/hosts | Update Discovery hosts (cloud infrastructure) |



## getConfigHosts

> CloudInfraConfig getConfigHosts()

Load Discovery hosts (cloud infrastructure)

### Example

```ts
import {
  Configuration,
  HostsApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigHostsRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new HostsApi();

  try {
    const data = await api.getConfigHosts();
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

[**CloudInfraConfig**](CloudInfraConfig.md)

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


## writeConfigHosts

> writeConfigHosts(cloudInfraConfig)

Update Discovery hosts (cloud infrastructure)

### Example

```ts
import {
  Configuration,
  HostsApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigHostsRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new HostsApi();

  const body = {
    // CloudInfraConfig
    cloudInfraConfig: ...,
  } satisfies WriteConfigHostsRequest;

  try {
    const data = await api.writeConfigHosts(body);
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
| **cloudInfraConfig** | [CloudInfraConfig](CloudInfraConfig.md) |  | |

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

