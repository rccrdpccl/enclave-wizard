# ClusterApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfigCluster**](ClusterApi.md#getconfigcluster) | **GET** /api/v1/config/cluster | Load Management cluster install configuration |
| [**writeConfigCluster**](ClusterApi.md#writeconfigcluster) | **PUT** /api/v1/config/cluster | Update Management cluster install configuration |



## getConfigCluster

> ClusterConfig getConfigCluster()

Load Management cluster install configuration

### Example

```ts
import {
  Configuration,
  ClusterApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigClusterRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new ClusterApi();

  try {
    const data = await api.getConfigCluster();
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

[**ClusterConfig**](ClusterConfig.md)

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


## writeConfigCluster

> writeConfigCluster(clusterConfig)

Update Management cluster install configuration

### Example

```ts
import {
  Configuration,
  ClusterApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigClusterRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new ClusterApi();

  const body = {
    // ClusterConfig
    clusterConfig: ...,
  } satisfies WriteConfigClusterRequest;

  try {
    const data = await api.writeConfigCluster(body);
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
| **clusterConfig** | [ClusterConfig](ClusterConfig.md) |  | |

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

