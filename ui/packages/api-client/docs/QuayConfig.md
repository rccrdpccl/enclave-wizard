
# QuayConfig


## Properties

Name | Type
------------ | -------------
`$schema` | string
`ocMirrorLogLevel` | string
`quayBackend` | string
`quayBackendRGWConfiguration` | [QuayBackendRGWConfiguration](QuayBackendRGWConfiguration.md)
`quayPassword` | string
`quayUser` | string

## Example

```typescript
import type { QuayConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "ocMirrorLogLevel": null,
  "quayBackend": null,
  "quayBackendRGWConfiguration": null,
  "quayPassword": null,
  "quayUser": null,
} satisfies QuayConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as QuayConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


