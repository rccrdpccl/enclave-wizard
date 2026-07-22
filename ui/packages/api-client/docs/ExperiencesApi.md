# ExperiencesApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**listExperiences**](ExperiencesApi.md#listexperiences) | **GET** /api/v1/experiences | List available experiences |



## listExperiences

> ExperiencesOutputBody listExperiences()

List available experiences

Returns experience definitions loaded from the enclave directory.

### Example

```ts
import {
  Configuration,
  ExperiencesApi,
} from '@enclave-wizard-ui/api-client';
import type { ListExperiencesRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new ExperiencesApi();

  try {
    const data = await api.listExperiences();
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

[**ExperiencesOutputBody**](ExperiencesOutputBody.md)

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

