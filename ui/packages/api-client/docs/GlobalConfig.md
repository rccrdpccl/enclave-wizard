
# GlobalConfig


## Properties

Name | Type
------------ | -------------
`agentHosts` | [Array&lt;HostEntry&gt;](HostEntry.md)
`apiVIP` | string
`baseDomain` | string
`blockStorageBackend` | string
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
`pullSecret` | any
`quayBackend` | string
`quayBackendRGWConfiguration` | [QuayBackendRGWConfiguration](QuayBackendRGWConfiguration.md)
`quayPassword` | string
`quayUser` | string
`rendezvousIP` | string
`sshPubPath` | string
`storagePlugin` | string
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
  "agentHosts": null,
  "apiVIP": null,
  "baseDomain": null,
  "blockStorageBackend": null,
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
  "pullSecret": null,
  "quayBackend": null,
  "quayBackendRGWConfiguration": null,
  "quayPassword": null,
  "quayUser": null,
  "rendezvousIP": null,
  "sshPubPath": null,
  "storagePlugin": null,
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


