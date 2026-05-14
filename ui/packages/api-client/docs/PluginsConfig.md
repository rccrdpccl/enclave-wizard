
# PluginsConfig


## Properties

Name | Type
------------ | -------------
`$schema` | string
`enabledPlugins` | Array&lt;string&gt;
`lvmsDefaults` | [LVMSConfig](LVMSConfig.md)
`odfDefaults` | [ODFConfig](ODFConfig.md)

## Example

```typescript
import type { PluginsConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "enabledPlugins": null,
  "lvmsDefaults": null,
  "odfDefaults": null,
} satisfies PluginsConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as PluginsConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


