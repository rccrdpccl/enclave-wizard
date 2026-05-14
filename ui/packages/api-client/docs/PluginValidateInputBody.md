
# PluginValidateInputBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`plugins` | Array&lt;string&gt;

## Example

```typescript
import type { PluginValidateInputBody } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "plugins": null,
} satisfies PluginValidateInputBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as PluginValidateInputBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


