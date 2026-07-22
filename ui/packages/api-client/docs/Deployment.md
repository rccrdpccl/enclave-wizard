
# Deployment


## Properties

Name | Type
------------ | -------------
`$schema` | string
`id` | string
`phases` | [Array&lt;DeploymentPhase&gt;](DeploymentPhase.md)
`startedAt` | Date
`status` | string
`totalTasks` | number

## Example

```typescript
import type { Deployment } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "id": null,
  "phases": null,
  "startedAt": null,
  "status": null,
  "totalTasks": null,
} satisfies Deployment

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as Deployment
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


