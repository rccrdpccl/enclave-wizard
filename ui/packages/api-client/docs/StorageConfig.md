
# StorageConfig


## Properties

Name | Type
------------ | -------------
`$schema` | string
`blockStorageBackend` | string
`lvmsConfig` | [LVMSStorageConfig](LVMSStorageConfig.md)
`odfExternalConfig` | string
`storagePlugin` | string
`vastAdminPassword` | string
`vastAdminUsername` | string
`vastEndpoint` | string
`vastVipPool` | [VASTVipPool](VASTVipPool.md)

## Example

```typescript
import type { StorageConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "blockStorageBackend": null,
  "lvmsConfig": null,
  "odfExternalConfig": null,
  "storagePlugin": null,
  "vastAdminPassword": null,
  "vastAdminUsername": null,
  "vastEndpoint": null,
  "vastVipPool": null,
} satisfies StorageConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as StorageConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


