
# EnclaveConfig


## Properties

Name | Type
------------ | -------------
`$schema` | string
`certificates` | [CertificatesConfig](CertificatesConfig.md)
`cloudInfra` | [CloudInfraConfig](CloudInfraConfig.md)
`global` | [GlobalConfig](GlobalConfig.md)

## Example

```typescript
import type { EnclaveConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "certificates": null,
  "cloudInfra": null,
  "global": null,
} satisfies EnclaveConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as EnclaveConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


