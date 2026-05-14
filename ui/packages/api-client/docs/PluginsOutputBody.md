
# PluginsOutputBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`plugins` | [Array&lt;Plugin&gt;](Plugin.md)

## Example

```typescript
import type { PluginsOutputBody } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "plugins": null,
} satisfies PluginsOutputBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as PluginsOutputBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


