
# GlobalConfig


## Properties

Name | Type
------------ | -------------
`aapDefaults` | [AAPConfig](AAPConfig.md)
`agentHosts` | [Array&lt;HostEntry&gt;](HostEntry.md)
`apiVIP` | string
`baseDomain` | string
`clusterFulfillmentConfig` | { [key: string]: string; }
`clusterName` | string
`defaultDNS` | string
`defaultGateway` | string
`defaultNtpServers` | Array&lt;string&gt;
`defaultPrefix` | number
`disconnected` | boolean
`diskEncryption` | boolean
`enabledPlugins` | Array&lt;string&gt;
`ingressVIP` | string
`lvmsConfig` | [LVMSStorageConfig](LVMSStorageConfig.md)
`lzBmcHostname` | string
`lzBmcIP` | string
`machineNetwork` | string
`masterMaxPods` | number
`ocMirrorLogLevel` | string
`odfDefaults` | [ODFConfig](ODFConfig.md)
`odfExternalConfig` | string
`osacAapLicenseFile` | string
`osacBYODatabase` | boolean
`osacBcmApiUrl` | string
`osacBcmClientCert` | string
`osacBcmClientKey` | string
`osacBcmDisableBmcCertVerification` | boolean
`osacBcmEnabled` | boolean
`osacBcmValidateCerts` | boolean
`osacDatabaseUrl` | string
`osacProfile` | string
`pullSecret` | any
`quayBackend` | string
`quayBackendRGWConfiguration` | [QuayBackendRGWConfiguration](QuayBackendRGWConfiguration.md)
`quayPassword` | string
`quayUser` | string
`rendezvousIP` | string
`rhbkDbSize` | string
`rhbkDeployDatabase` | boolean
`rhbkInstances` | number
`sshPubKey` | string
`storagePlugin` | string
`trustManagerDefaults` | [TrustManagerConfig](TrustManagerConfig.md)
`vastAdminPassword` | string
`vastAdminUsername` | string
`vastDefaults` | [VASTConfig](VASTConfig.md)
`vastEndpoint` | string
`vastVipPool` | [VASTVipPool](VASTVipPool.md)
`workingDir` | string

## Example

```typescript
import type { GlobalConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "aapDefaults": null,
  "agentHosts": null,
  "apiVIP": null,
  "baseDomain": null,
  "clusterFulfillmentConfig": null,
  "clusterName": null,
  "defaultDNS": null,
  "defaultGateway": null,
  "defaultNtpServers": null,
  "defaultPrefix": null,
  "disconnected": null,
  "diskEncryption": null,
  "enabledPlugins": null,
  "ingressVIP": null,
  "lvmsConfig": null,
  "lzBmcHostname": null,
  "lzBmcIP": null,
  "machineNetwork": null,
  "masterMaxPods": null,
  "ocMirrorLogLevel": null,
  "odfDefaults": null,
  "odfExternalConfig": null,
  "osacAapLicenseFile": null,
  "osacBYODatabase": null,
  "osacBcmApiUrl": null,
  "osacBcmClientCert": null,
  "osacBcmClientKey": null,
  "osacBcmDisableBmcCertVerification": null,
  "osacBcmEnabled": null,
  "osacBcmValidateCerts": null,
  "osacDatabaseUrl": null,
  "osacProfile": null,
  "pullSecret": null,
  "quayBackend": null,
  "quayBackendRGWConfiguration": null,
  "quayPassword": null,
  "quayUser": null,
  "rendezvousIP": null,
  "rhbkDbSize": null,
  "rhbkDeployDatabase": null,
  "rhbkInstances": null,
  "sshPubKey": null,
  "storagePlugin": null,
  "trustManagerDefaults": null,
  "vastAdminPassword": null,
  "vastAdminUsername": null,
  "vastDefaults": null,
  "vastEndpoint": null,
  "vastVipPool": null,
  "workingDir": null,
} satisfies GlobalConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as GlobalConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


