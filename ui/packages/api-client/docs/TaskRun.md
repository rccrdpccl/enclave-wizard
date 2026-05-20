
# TaskRun


## Properties

Name | Type
------------ | -------------
`$schema` | string
`endedAt` | Date
`error` | string
`exitCode` | number
`extraVars` | { [key: string]: string; }
`id` | string
`pid` | number
`playbook` | string
`startedAt` | Date
`status` | string
`type` | string

## Example

```typescript
import type { TaskRun } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "endedAt": null,
  "error": null,
  "exitCode": null,
  "extraVars": null,
  "id": null,
  "pid": null,
  "playbook": null,
  "startedAt": null,
  "status": null,
  "type": null,
} satisfies TaskRun

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as TaskRun
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


