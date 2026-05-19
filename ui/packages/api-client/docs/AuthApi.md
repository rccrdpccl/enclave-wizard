# AuthApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**changePassword**](AuthApi.md#changepassword) | **POST** /api/v1/auth/password | Change admin password |
| [**login**](AuthApi.md#login) | **POST** /api/v1/auth/login | Authenticate and get a bearer token |



## changePassword

> ChangePasswordOutputBody changePassword(changePasswordInputBody)

Change admin password

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '@enclave-wizard-ui/api-client';
import type { ChangePasswordRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new AuthApi(config);

  const body = {
    // ChangePasswordInputBody
    changePasswordInputBody: ...,
  } satisfies ChangePasswordRequest;

  try {
    const data = await api.changePassword(body);
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
| **changePasswordInputBody** | [ChangePasswordInputBody](ChangePasswordInputBody.md) |  | |

### Return type

[**ChangePasswordOutputBody**](ChangePasswordOutputBody.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## login

> LoginOutputBody login(loginInputBody)

Authenticate and get a bearer token

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '@enclave-wizard-ui/api-client';
import type { LoginRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new AuthApi();

  const body = {
    // LoginInputBody
    loginInputBody: ...,
  } satisfies LoginRequest;

  try {
    const data = await api.login(body);
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
| **loginInputBody** | [LoginInputBody](LoginInputBody.md) |  | |

### Return type

[**LoginOutputBody**](LoginOutputBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

