
# Defaults


## Properties

Name | Type
------------ | -------------
`$schema` | string
`disconnected` | boolean
`diskEncryption` | boolean
`masterMaxPods` | number
`ocMirrorLogLevel` | string
`pluginDefaults` | { [key: string]: any; }
`storagePlugin` | string

## Example

```typescript
import type { Defaults } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "disconnected": null,
  "diskEncryption": null,
  "masterMaxPods": null,
  "ocMirrorLogLevel": null,
  "pluginDefaults": null,
  "storagePlugin": null,
} satisfies Defaults

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as Defaults
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


