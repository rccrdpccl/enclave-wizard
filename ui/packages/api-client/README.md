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
import type { AuthModeRequest } from '@enclave-wizard-ui/api-client';

async function example() {
  console.log("🚀 Testing @enclave-wizard-ui/api-client SDK...");
  const api = new AuthApi();

  try {
    const data = await api.authMode();
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
*AuthApi* | [**authMode**](docs/AuthApi.md#authmode) | **GET** /api/v1/auth/mode | Check authentication mode
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
*DeploymentsApi* | [**cancelDeployment**](docs/DeploymentsApi.md#canceldeployment) | **DELETE** /api/v1/deployments/{id} | Cancel deployment
*DeploymentsApi* | [**getCurrentDeployment**](docs/DeploymentsApi.md#getcurrentdeployment) | **GET** /api/v1/deployments/current | Get current deployment
*DeploymentsApi* | [**getDeploymentById**](docs/DeploymentsApi.md#getdeploymentbyid) | **GET** /api/v1/deployments/{id} | Get deployment by ID
*DeploymentsApi* | [**getDeploymentProgressById**](docs/DeploymentsApi.md#getdeploymentprogressbyid) | **GET** /api/v1/deployments/{id}/progress | Get deployment progress
*DeploymentsApi* | [**startDeployment**](docs/DeploymentsApi.md#startdeployment) | **POST** /api/v1/deployments | Start deployment
*ExperiencesApi* | [**listExperiences**](docs/ExperiencesApi.md#listexperiences) | **GET** /api/v1/experiences | List available experiences
*FilesApi* | [**uploadFile**](docs/FilesApi.md#uploadfile) | **POST** /api/v1/files | Upload a file
*HostsApi* | [**getConfigHosts**](docs/HostsApi.md#getconfighosts) | **GET** /api/v1/config/hosts | Load Discovery hosts (cloud infrastructure)
*HostsApi* | [**writeConfigHosts**](docs/HostsApi.md#writeconfighosts) | **PUT** /api/v1/config/hosts | Update Discovery hosts (cloud infrastructure)
*LandingZoneApi* | [**getConfigLz**](docs/LandingZoneApi.md#getconfiglz) | **GET** /api/v1/config/lz | Load Landing zone configuration
*LandingZoneApi* | [**writeConfigLz**](docs/LandingZoneApi.md#writeconfiglz) | **PUT** /api/v1/config/lz | Update Landing zone configuration
*NetworkApi* | [**getConfigNetwork**](docs/NetworkApi.md#getconfignetwork) | **GET** /api/v1/config/network | Load Host network configuration
*NetworkApi* | [**writeConfigNetwork**](docs/NetworkApi.md#writeconfignetwork) | **PUT** /api/v1/config/network | Update Host network configuration
*PluginsApi* | [**getConfigPlugins**](docs/PluginsApi.md#getconfigplugins) | **GET** /api/v1/config/plugins | Load Enabled plugins configuration
*PluginsApi* | [**getPluginSchema**](docs/PluginsApi.md#getpluginschema) | **GET** /api/v1/plugins/{name}/schema | Get plugin config schema
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
*TasksApi* | [**startValidate**](docs/TasksApi.md#startvalidate) | **POST** /api/v1/tasks/validate | Run operational validation (validations.sh)
*VersionApi* | [**getVersion**](docs/VersionApi.md#getversion) | **GET** /api/v1/version | Get wizard and enclave versions


### Models

- [AAPConfig](docs/AAPConfig.md)
- [AuthModeOutputBody](docs/AuthModeOutputBody.md)
- [CertificatesConfig](docs/CertificatesConfig.md)
- [ChangePasswordInputBody](docs/ChangePasswordInputBody.md)
- [ChangePasswordOutputBody](docs/ChangePasswordOutputBody.md)
- [CloudInfraConfig](docs/CloudInfraConfig.md)
- [ClusterConfig](docs/ClusterConfig.md)
- [Defaults](docs/Defaults.md)
- [Deployment](docs/Deployment.md)
- [DeploymentPhase](docs/DeploymentPhase.md)
- [DeploymentProgress](docs/DeploymentProgress.md)
- [EnclaveConfig](docs/EnclaveConfig.md)
- [ErrorDetail](docs/ErrorDetail.md)
- [ErrorModel](docs/ErrorModel.md)
- [Experience](docs/Experience.md)
- [ExperiencePlugin](docs/ExperiencePlugin.md)
- [ExperiencesOutputBody](docs/ExperiencesOutputBody.md)
- [FileUploadOutputBody](docs/FileUploadOutputBody.md)
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
- [TrustManagerConfig](docs/TrustManagerConfig.md)
- [VASTConfig](docs/VASTConfig.md)
- [VASTIPRange](docs/VASTIPRange.md)
- [VASTTier](docs/VASTTier.md)
- [VASTVipPool](docs/VASTVipPool.md)
- [ValidateConfigOutputBody](docs/ValidateConfigOutputBody.md)
- [ValidationError](docs/ValidationError.md)
- [VersionOutputBody](docs/VersionOutputBody.md)

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
