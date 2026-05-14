# LandingZoneApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getConfigLz**](LandingZoneApi.md#getconfiglz) | **GET** /api/v1/config/lz | Load Landing zone configuration |
| [**writeConfigLz**](LandingZoneApi.md#writeconfiglz) | **PUT** /api/v1/config/lz | Update Landing zone configuration |



## getConfigLz

> LandingZoneConfig getConfigLz()

Load Landing zone configuration

### Example

```ts
import {
  Configuration,
  LandingZoneApi,
} from '@enclave-wizard-ui/api-client';
import type { GetConfigLzRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new LandingZoneApi();

  try {
    const data = await api.getConfigLz();
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

[**LandingZoneConfig**](LandingZoneConfig.md)

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


## writeConfigLz

> writeConfigLz(landingZoneConfig)

Update Landing zone configuration

### Example

```ts
import {
  Configuration,
  LandingZoneApi,
} from '@enclave-wizard-ui/api-client';
import type { WriteConfigLzRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new LandingZoneApi();

  const body = {
    // LandingZoneConfig
    landingZoneConfig: ...,
  } satisfies WriteConfigLzRequest;

  try {
    const data = await api.writeConfigLz(body);
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
| **landingZoneConfig** | [LandingZoneConfig](LandingZoneConfig.md) |  | |

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

