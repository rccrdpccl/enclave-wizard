# @enclave-wizard-ui/api-client@0.1.0

A TypeScript SDK client for the localhost API.

## Usage

First, install the SDK from npm.

```bash
npm install @enclave-wizard-ui/api-client --save
```

Next, try it out.


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


## Documentation

### API Endpoints

All URIs are relative to *http://localhost*

| Class | Method | HTTP request | Description
| ----- | ------ | ------------ | -------------
*AuthApi* | [**changePassword**](docs/AuthApi.md#changepassword) | **POST** /api/v1/auth/password | Change admin password
*AuthApi* | [**login**](docs/AuthApi.md#login) | **POST** /api/v1/auth/login | Authenticate and get a bearer token
*CertificatesApi* | [**getConfigCertificates**](docs/CertificatesApi.md#getconfigcertificates) | **GET** /api/v1/config/certificates | Load TLS certificates
*CertificatesApi* | [**writeConfigCertificates**](docs/CertificatesApi.md#writeconfigcertificates) | **PUT** /api/v1/config/certificates | Update TLS certificates
*ClusterApi* | [**getConfigCluster**](docs/ClusterApi.md#getconfigcluster) | **GET** /api/v1/config/cluster | Load Management cluster install configuration
*ClusterApi* | [**writeConfigCluster**](docs/ClusterApi.md#writeconfigcluster) | **PUT** /api/v1/config/cluster | Update Management cluster install configuration
*ConfigApi* | [**getConfig**](docs/ConfigApi.md#getconfig) | **GET** /api/v1/config | Load existing configuration
*ConfigApi* | [**previewConfig**](docs/ConfigApi.md#previewconfig) | **POST** /api/v1/config/preview | Preview rendered YAML
*ConfigApi* | [**validateConfig**](docs/ConfigApi.md#validateconfig) | **POST** /api/v1/config/validate | Validate configuration
*ConfigApi* | [**writeConfig**](docs/ConfigApi.md#writeconfig) | **PUT** /api/v1/config | Write configuration to disk
*DefaultsApi* | [**getDefaults**](docs/DefaultsApi.md#getdefaults) | **GET** /api/v1/defaults | Get default configuration values
*HostsApi* | [**getConfigHosts**](docs/HostsApi.md#getconfighosts) | **GET** /api/v1/config/hosts | Load Discovery hosts (cloud infrastructure)
*HostsApi* | [**writeConfigHosts**](docs/HostsApi.md#writeconfighosts) | **PUT** /api/v1/config/hosts | Update Discovery hosts (cloud infrastructure)
*LandingZoneApi* | [**getConfigLz**](docs/LandingZoneApi.md#getconfiglz) | **GET** /api/v1/config/lz | Load Landing zone configuration
*LandingZoneApi* | [**writeConfigLz**](docs/LandingZoneApi.md#writeconfiglz) | **PUT** /api/v1/config/lz | Update Landing zone configuration
*NetworkApi* | [**getConfigNetwork**](docs/NetworkApi.md#getconfignetwork) | **GET** /api/v1/config/network | Load Host network configuration
*NetworkApi* | [**writeConfigNetwork**](docs/NetworkApi.md#writeconfignetwork) | **PUT** /api/v1/config/network | Update Host network configuration
*PluginsApi* | [**getConfigPlugins**](docs/PluginsApi.md#getconfigplugins) | **GET** /api/v1/config/plugins | Load Enabled plugins configuration
*PluginsApi* | [**listPlugins**](docs/PluginsApi.md#listplugins) | **GET** /api/v1/plugins | List available plugins
*PluginsApi* | [**validatePluginCombination**](docs/PluginsApi.md#validateplugincombination) | **POST** /api/v1/plugins/validate | Validate plugin combination
*PluginsApi* | [**writeConfigPlugins**](docs/PluginsApi.md#writeconfigplugins) | **PUT** /api/v1/config/plugins | Update Enabled plugins configuration
*QuayApi* | [**getConfigQuay**](docs/QuayApi.md#getconfigquay) | **GET** /api/v1/config/quay | Load Quay registry configuration
*QuayApi* | [**writeConfigQuay**](docs/QuayApi.md#writeconfigquay) | **PUT** /api/v1/config/quay | Update Quay registry configuration
*StorageApi* | [**getConfigStorage**](docs/StorageApi.md#getconfigstorage) | **GET** /api/v1/config/storage | Load Block storage configuration
*StorageApi* | [**writeConfigStorage**](docs/StorageApi.md#writeconfigstorage) | **PUT** /api/v1/config/storage | Update Block storage configuration
*TasksApi* | [**deleteTask**](docs/TasksApi.md#deletetask) | **DELETE** /api/v1/tasks/{id} | Delete a task run
*TasksApi* | [**getTask**](docs/TasksApi.md#gettask) | **GET** /api/v1/tasks/{id} | Get task run details
*TasksApi* | [**getTaskEvents**](docs/TasksApi.md#gettaskevents) | **GET** /api/v1/tasks/{id}/events | Get task job events
*TasksApi* | [**getTaskLogs**](docs/TasksApi.md#gettasklogs) | **GET** /api/v1/tasks/{id}/logs | Get task output logs
*TasksApi* | [**listTasks**](docs/TasksApi.md#listtasks) | **GET** /api/v1/tasks | List all task runs
*TasksApi* | [**startDeploy**](docs/TasksApi.md#startdeploy) | **POST** /api/v1/tasks/deploy | Start full deployment
*TasksApi* | [**startDeployPhase**](docs/TasksApi.md#startdeployphase) | **POST** /api/v1/tasks/deploy/{phase} | Start a specific deployment phase
*TasksApi* | [**startDeployPlugin**](docs/TasksApi.md#startdeployplugin) | **POST** /api/v1/tasks/plugins/{name} | Deploy a plugin


### Models

- [CertificatesConfig](docs/CertificatesConfig.md)
- [ChangePasswordInputBody](docs/ChangePasswordInputBody.md)
- [ChangePasswordOutputBody](docs/ChangePasswordOutputBody.md)
- [CloudInfraConfig](docs/CloudInfraConfig.md)
- [ClusterConfig](docs/ClusterConfig.md)
- [Defaults](docs/Defaults.md)
- [EnclaveConfig](docs/EnclaveConfig.md)
- [ErrorDetail](docs/ErrorDetail.md)
- [ErrorModel](docs/ErrorModel.md)
- [GetTaskEventsOutputBody](docs/GetTaskEventsOutputBody.md)
- [GlobalConfig](docs/GlobalConfig.md)
- [HostEntry](docs/HostEntry.md)
- [LVMSConfig](docs/LVMSConfig.md)
- [LVMSDeviceSelector](docs/LVMSDeviceSelector.md)
- [LVMSStorageConfig](docs/LVMSStorageConfig.md)
- [LVMSThinPoolConfig](docs/LVMSThinPoolConfig.md)
- [LandingZoneConfig](docs/LandingZoneConfig.md)
- [ListTasksOutputBody](docs/ListTasksOutputBody.md)
- [LoginInputBody](docs/LoginInputBody.md)
- [LoginOutputBody](docs/LoginOutputBody.md)
- [NetworkConfig](docs/NetworkConfig.md)
- [ODFConfig](docs/ODFConfig.md)
- [Plugin](docs/Plugin.md)
- [PluginValidateInputBody](docs/PluginValidateInputBody.md)
- [PluginValidateOutputBody](docs/PluginValidateOutputBody.md)
- [PluginsConfig](docs/PluginsConfig.md)
- [PluginsOutputBody](docs/PluginsOutputBody.md)
- [PreviewConfigOutputBody](docs/PreviewConfigOutputBody.md)
- [QuayBackendRGWConfiguration](docs/QuayBackendRGWConfiguration.md)
- [QuayConfig](docs/QuayConfig.md)
- [StorageConfig](docs/StorageConfig.md)
- [TaskRun](docs/TaskRun.md)
- [VASTConfig](docs/VASTConfig.md)
- [VASTIPRange](docs/VASTIPRange.md)
- [VASTTier](docs/VASTTier.md)
- [VASTVipPool](docs/VASTVipPool.md)
- [ValidateConfigOutputBody](docs/ValidateConfigOutputBody.md)
- [ValidationError](docs/ValidationError.md)

### Authorization


Authentication schemes defined for the API:
<a id="bearer"></a>
#### bearer


- **Type**: HTTP Bearer Token authentication (opaque)

## About

This TypeScript SDK client supports the [Fetch API](https://fetch.spec.whatwg.org/)
and is automatically generated by the
[OpenAPI Generator](https://openapi-generator.tech) project:

- API version: `0.1.0`
- Package version: `0.1.0`
- Generator version: `7.18.0`
- Build package: `org.openapitools.codegen.languages.TypeScriptFetchClientCodegen`

The generated npm module supports the following:

- Environments
  * Node.js
  * Webpack
  * Browserify
- Language levels
  * ES5 - you must have a Promises/A+ library installed
  * ES6
- Module systems
  * CommonJS
  * ES6 module system


## Development

### Building

To build the TypeScript source code, you need to have Node.js and npm installed.
After cloning the repository, navigate to the project directory and run:

```bash
npm install
npm run build
```

### Publishing

Once you've built the package, you can publish it to npm:

```bash
npm publish
```

## License

[]()
