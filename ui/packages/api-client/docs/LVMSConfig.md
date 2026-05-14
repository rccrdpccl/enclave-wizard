
# LVMSConfig


## Properties

Name | Type
------------ | -------------
`defaultStorageClass` | boolean
`deviceClassName` | string
`thinPoolConfig` | [LVMSThinPoolConfig](LVMSThinPoolConfig.md)

## Example

```typescript
import type { LVMSConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "defaultStorageClass": null,
  "deviceClassName": null,
  "thinPoolConfig": null,
} satisfies LVMSConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as LVMSConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


