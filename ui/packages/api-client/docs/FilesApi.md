# FilesApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**uploadFile**](FilesApi.md#uploadfile) | **POST** /api/v1/files | Upload a file |



## uploadFile

> FileUploadOutputBody uploadFile(filename, name)

Upload a file

Upload a file to the enclave config directory. The file is written to config/&lt;dest&gt;/&lt;filename&gt;.

### Example

```ts
import {
  Configuration,
  FilesApi,
} from '@enclave-wizard-ui/api-client';
import type { UploadFileRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new FilesApi();

  const body = {
    // Blob | filename of the file being uploaded (optional)
    filename: BINARY_DATA_HERE,
    // string | general purpose name for multipart form value (optional)
    name: name_example,
  } satisfies UploadFileRequest;

  try {
    const data = await api.uploadFile(body);
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
| **filename** | `Blob` | filename of the file being uploaded | [Optional] [Defaults to `undefined`] |
| **name** | `string` | general purpose name for multipart form value | [Optional] [Defaults to `undefined`] |

### Return type

[**FileUploadOutputBody**](FileUploadOutputBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `multipart/form-data`
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

