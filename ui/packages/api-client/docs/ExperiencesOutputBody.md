
# ExperiencesOutputBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`experiences` | [Array&lt;Experience&gt;](Experience.md)

## Example

```typescript
import type { ExperiencesOutputBody } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "experiences": null,
} satisfies ExperiencesOutputBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ExperiencesOutputBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


