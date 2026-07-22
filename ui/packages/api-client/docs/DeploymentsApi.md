# DeploymentsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**cancelDeployment**](DeploymentsApi.md#canceldeployment) | **DELETE** /api/v1/deployments/{id} | Cancel deployment |
| [**getCurrentDeployment**](DeploymentsApi.md#getcurrentdeployment) | **GET** /api/v1/deployments/current | Get current deployment |
| [**getDeploymentById**](DeploymentsApi.md#getdeploymentbyid) | **GET** /api/v1/deployments/{id} | Get deployment by ID |
| [**getDeploymentProgressById**](DeploymentsApi.md#getdeploymentprogressbyid) | **GET** /api/v1/deployments/{id}/progress | Get deployment progress |
| [**startDeployment**](DeploymentsApi.md#startdeployment) | **POST** /api/v1/deployments | Start deployment |



## cancelDeployment

> cancelDeployment(id)

Cancel deployment

Cancels a running deployment.

### Example

```ts
import {
  Configuration,
  DeploymentsApi,
} from '@enclave-wizard-ui/api-client';
import type { CancelDeploymentRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new DeploymentsApi();

  const body = {
    // string | Deployment identifier
    id: id_example,
  } satisfies CancelDeploymentRequest;

  try {
    const data = await api.cancelDeployment(body);
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
| **id** | `string` | Deployment identifier | [Defaults to `undefined`] |

### Return type

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getCurrentDeployment

> Deployment getCurrentDeployment()

Get current deployment

Returns the most recent deployment. Use this for reconnection after page reload.

### Example

```ts
import {
  Configuration,
  DeploymentsApi,
} from '@enclave-wizard-ui/api-client';
import type { GetCurrentDeploymentRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new DeploymentsApi();

  try {
    const data = await api.getCurrentDeployment();
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

[**Deployment**](Deployment.md)

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


## getDeploymentById

> Deployment getDeploymentById(id)

Get deployment by ID

Returns the deployment state with phases for the given deployment ID.

### Example

```ts
import {
  Configuration,
  DeploymentsApi,
} from '@enclave-wizard-ui/api-client';
import type { GetDeploymentByIdRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new DeploymentsApi();

  const body = {
    // string | Deployment identifier
    id: id_example,
  } satisfies GetDeploymentByIdRequest;

  try {
    const data = await api.getDeploymentById(body);
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
| **id** | `string` | Deployment identifier | [Defaults to `undefined`] |

### Return type

[**Deployment**](Deployment.md)

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


## getDeploymentProgressById

> DeploymentProgress getDeploymentProgressById(id)

Get deployment progress

Returns live progress with completed task count, percentage, and current phase/task.

### Example

```ts
import {
  Configuration,
  DeploymentsApi,
} from '@enclave-wizard-ui/api-client';
import type { GetDeploymentProgressByIdRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new DeploymentsApi();

  const body = {
    // string | Deployment identifier
    id: id_example,
  } satisfies GetDeploymentProgressByIdRequest;

  try {
    const data = await api.getDeploymentProgressById(body);
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
| **id** | `string` | Deployment identifier | [Defaults to `undefined`] |

### Return type

[**DeploymentProgress**](DeploymentProgress.md)

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


## startDeployment

> Deployment startDeployment()

Start deployment

Starts the full deployment chain (main playbook + addon plugins).

### Example

```ts
import {
  Configuration,
  DeploymentsApi,
} from '@enclave-wizard-ui/api-client';
import type { StartDeploymentRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new DeploymentsApi();

  try {
    const data = await api.startDeployment();
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

[**Deployment**](Deployment.md)

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

