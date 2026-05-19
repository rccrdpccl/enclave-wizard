
# VASTConfig


## Properties

Name | Type
------------ | -------------
`infraTenant` | string
`quayPvcSize` | string
`storagePath` | string
`tiers` | [Array&lt;VASTTier&gt;](VASTTier.md)
`viewPolicyId` | number

## Example

```typescript
import type { VASTConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "infraTenant": null,
  "quayPvcSize": null,
  "storagePath": null,
  "tiers": null,
  "viewPolicyId": null,
} satisfies VASTConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as VASTConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


