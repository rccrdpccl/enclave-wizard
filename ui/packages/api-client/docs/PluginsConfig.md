
# PluginsConfig


## Properties

Name | Type
------------ | -------------
`$schema` | string
`aapDefaults` | [AAPConfig](AAPConfig.md)
`clusterFulfillmentConfig` | { [key: string]: string; }
`enabledPlugins` | Array&lt;string&gt;
`lvmsDefaults` | [LVMSConfig](LVMSConfig.md)
`odfDefaults` | [ODFConfig](ODFConfig.md)
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
`rhbkDbSize` | string
`rhbkDeployDatabase` | boolean
`rhbkInstances` | number
`trustManagerDefaults` | [TrustManagerConfig](TrustManagerConfig.md)
`vastDefaults` | [VASTConfig](VASTConfig.md)

## Example

```typescript
import type { PluginsConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "aapDefaults": null,
  "clusterFulfillmentConfig": null,
  "enabledPlugins": null,
  "lvmsDefaults": null,
  "odfDefaults": null,
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
  "rhbkDbSize": null,
  "rhbkDeployDatabase": null,
  "rhbkInstances": null,
  "trustManagerDefaults": null,
  "vastDefaults": null,
} satisfies PluginsConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as PluginsConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


