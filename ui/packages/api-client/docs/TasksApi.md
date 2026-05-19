# TasksApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getTask**](TasksApi.md#gettask) | **GET** /api/v1/tasks/{id} | Get task run details |
| [**getTaskEvents**](TasksApi.md#gettaskevents) | **GET** /api/v1/tasks/{id}/events | Get task job events |
| [**getTaskLogs**](TasksApi.md#gettasklogs) | **GET** /api/v1/tasks/{id}/logs | Get task output logs |
| [**listTasks**](TasksApi.md#listtasks) | **GET** /api/v1/tasks | List all task runs |
| [**startDeploy**](TasksApi.md#startdeploy) | **POST** /api/v1/tasks/deploy | Start full deployment |
| [**startDeployPhase**](TasksApi.md#startdeployphase) | **POST** /api/v1/tasks/deploy/{phase} | Start a specific deployment phase |
| [**startDeployPlugin**](TasksApi.md#startdeployplugin) | **POST** /api/v1/tasks/plugins/{name} | Deploy a plugin |



## getTask

> TaskRun getTask(id)

Get task run details

Returns status and metadata for a specific run.

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '@enclave-wizard-ui/api-client';
import type { GetTaskRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new TasksApi();

  const body = {
    // string | Run identifier
    id: id_example,
  } satisfies GetTaskRequest;

  try {
    const data = await api.getTask(body);
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
| **id** | `string` | Run identifier | [Defaults to `undefined`] |

### Return type

[**TaskRun**](TaskRun.md)

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


## getTaskEvents

> GetTaskEventsOutputBody getTaskEvents(id)

Get task job events

Returns ansible-runner job events as a JSON array.

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '@enclave-wizard-ui/api-client';
import type { GetTaskEventsRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new TasksApi();

  const body = {
    // string | Run identifier
    id: id_example,
  } satisfies GetTaskEventsRequest;

  try {
    const data = await api.getTaskEvents(body);
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
| **id** | `string` | Run identifier | [Defaults to `undefined`] |

### Return type

[**GetTaskEventsOutputBody**](GetTaskEventsOutputBody.md)

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


## getTaskLogs

> string getTaskLogs(id)

Get task output logs

Returns ansible-runner stdout as text/plain. Use the offset query parameter for incremental reads.

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '@enclave-wizard-ui/api-client';
import type { GetTaskLogsRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new TasksApi();

  const body = {
    // string | Run identifier
    id: id_example,
  } satisfies GetTaskLogsRequest;

  try {
    const data = await api.getTaskLogs(body);
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
| **id** | `string` | Run identifier | [Defaults to `undefined`] |

### Return type

**string**

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


## listTasks

> ListTasksOutputBody listTasks()

List all task runs

Returns all known task runs, most recent first.

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '@enclave-wizard-ui/api-client';
import type { ListTasksRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new TasksApi();

  try {
    const data = await api.listTasks();
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

[**ListTasksOutputBody**](ListTasksOutputBody.md)

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


## startDeploy

> TaskRun startDeploy()

Start full deployment

Runs the main.yaml playbook (all 7 phases).

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '@enclave-wizard-ui/api-client';
import type { StartDeployRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new TasksApi();

  try {
    const data = await api.startDeploy();
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

[**TaskRun**](TaskRun.md)

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


## startDeployPhase

> TaskRun startDeployPhase(phase)

Start a specific deployment phase

Runs a single deployment phase (1-7).

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '@enclave-wizard-ui/api-client';
import type { StartDeployPhaseRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new TasksApi();

  const body = {
    // number | Deployment phase number (1-7)
    phase: 789,
  } satisfies StartDeployPhaseRequest;

  try {
    const data = await api.startDeployPhase(body);
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
| **phase** | `number` | Deployment phase number (1-7) | [Defaults to `undefined`] |

### Return type

[**TaskRun**](TaskRun.md)

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


## startDeployPlugin

> TaskRun startDeployPlugin(name)

Deploy a plugin

Runs the deploy-plugin.yaml playbook for the named plugin.

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '@enclave-wizard-ui/api-client';
import type { StartDeployPluginRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new TasksApi();

  const body = {
    // string | Plugin name
    name: name_example,
  } satisfies StartDeployPluginRequest;

  try {
    const data = await api.startDeployPlugin(body);
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
| **name** | `string` | Plugin name | [Defaults to `undefined`] |

### Return type

[**TaskRun**](TaskRun.md)

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

