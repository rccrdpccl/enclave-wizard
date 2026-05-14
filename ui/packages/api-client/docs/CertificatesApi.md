# CertificatesApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfigCertificates**](CertificatesApi.md#getconfigcertificates) | **GET** /api/v1/config/certificates | Load TLS certificates |
| [**writeConfigCertificates**](CertificatesApi.md#writeconfigcertificates) | **PUT** /api/v1/config/certificates | Update TLS certificates |



## getConfigCertificates

> CertificatesConfig getConfigCertificates()

Load TLS certificates

### Example

```ts
import {
  Configuration,
  CertificatesApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigCertificatesRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new CertificatesApi();

  try {
    const data = await api.getConfigCertificates();
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

[**CertificatesConfig**](CertificatesConfig.md)

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


## writeConfigCertificates

> writeConfigCertificates(certificatesConfig)

Update TLS certificates

### Example

```ts
import {
  Configuration,
  CertificatesApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigCertificatesRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new CertificatesApi();

  const body = {
    // CertificatesConfig
    certificatesConfig: ...,
  } satisfies WriteConfigCertificatesRequest;

  try {
    const data = await api.writeConfigCertificates(body);
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
| **certificatesConfig** | [CertificatesConfig](CertificatesConfig.md) |  | |

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

