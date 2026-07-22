
# TrustManagerConfig


## Properties

Name | Type
------------ | -------------
`trustManagerCaIssuerDuration` | string
`trustManagerCaIssuerRenewBefore` | string

## Example

```typescript
import type { TrustManagerConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "trustManagerCaIssuerDuration": null,
  "trustManagerCaIssuerRenewBefore": null,
} satisfies TrustManagerConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as TrustManagerConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


