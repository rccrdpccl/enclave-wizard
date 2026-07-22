
# DeploymentProgress


## Properties

Name | Type
------------ | -------------
`$schema` | string
`completed` | number
`currentPhase` | string
`currentTask` | string
`percentage` | number
`total` | number

## Example

```typescript
import type { DeploymentProgress } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "completed": null,
  "currentPhase": null,
  "currentTask": null,
  "percentage": null,
  "total": null,
} satisfies DeploymentProgress

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as DeploymentProgress
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


