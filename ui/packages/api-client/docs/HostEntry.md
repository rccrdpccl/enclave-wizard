
# HostEntry


## Properties

Name | Type
------------ | -------------
`bmcSystemId` | string
`ipAddress` | string
`macAddress` | string
`mapInterfaces` | any
`name` | string
`networkConfig` | any
`redfish` | string
`redfishPassword` | string
`redfishUser` | string
`rootDisk` | string

## Example

```typescript
import type { HostEntry } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "bmcSystemId": null,
  "ipAddress": null,
  "macAddress": null,
  "mapInterfaces": null,
  "name": null,
  "networkConfig": null,
  "redfish": null,
  "redfishPassword": null,
  "redfishUser": null,
  "rootDisk": null,
} satisfies HostEntry

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as HostEntry
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


