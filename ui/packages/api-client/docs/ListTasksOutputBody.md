
# ListTasksOutputBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`runs` | [Array&lt;TaskRun&gt;](TaskRun.md)

## Example

```typescript
import type { ListTasksOutputBody } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "runs": null,
} satisfies ListTasksOutputBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ListTasksOutputBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


