
# LandingZoneConfig


## Properties

Name | Type
------------ | -------------
`$schema` | string
`disconnected` | boolean
`lzBmcHostname` | string
`lzBmcIP` | string
`workingDir` | string

## Example

```typescript
import type { LandingZoneConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "disconnected": null,
  "lzBmcHostname": null,
  "lzBmcIP": null,
  "workingDir": null,
} satisfies LandingZoneConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as LandingZoneConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


