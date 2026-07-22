
# AAPConfig


## Properties

Name | Type
------------ | -------------
`aapLicenseFile` | string
`aapControllerDisabled` | boolean
`aapEdaDisabled` | boolean
`aapHubDisabled` | boolean
`aapImagePullPolicy` | string
`aapLicenseSecret` | string
`aapLightspeedDisabled` | boolean
`aapName` | string
`aapNoLog` | boolean
`aapNs` | string
`aapRedisMode` | string
`aapRouteTlsTermination` | string

## Example

```typescript
import type { AAPConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "aapLicenseFile": null,
  "aapControllerDisabled": null,
  "aapEdaDisabled": null,
  "aapHubDisabled": null,
  "aapImagePullPolicy": null,
  "aapLicenseSecret": null,
  "aapLightspeedDisabled": null,
  "aapName": null,
  "aapNoLog": null,
  "aapNs": null,
  "aapRedisMode": null,
  "aapRouteTlsTermination": null,
} satisfies AAPConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as AAPConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


