
# LVMSThinPoolConfig


## Properties

Name | Type
------------ | -------------
`name` | string
`overprovisionRatio` | number
`sizePercent` | number

## Example

```typescript
import type { LVMSThinPoolConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "name": null,
  "overprovisionRatio": null,
  "sizePercent": null,
} satisfies LVMSThinPoolConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as LVMSThinPoolConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


