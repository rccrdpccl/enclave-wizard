
# TaskRun


## Properties

Name | Type
------------ | -------------
`id` | string
`type` | [TaskType](TaskType.md)
`status` | [TaskStatus](TaskStatus.md)
`playbook` | string
`extraVars` | { [key: string]: string; }
`pid` | number
`exitCode` | number
`createdAt` | Date
`startedAt` | Date
`endedAt` | Date
`error` | string

## Example

```typescript
import type { TaskRun } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "id": null,
  "type": null,
  "status": null,
  "playbook": null,
  "extraVars": null,
  "pid": null,
  "exitCode": null,
  "createdAt": null,
  "startedAt": null,
  "endedAt": null,
  "error": null,
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


