
# QuayBackendRGWConfiguration


## Properties

Name | Type
------------ | -------------
`accessKey` | string
`bucketName` | string
`hostname` | string
`isSecure` | boolean
`maximumChunkSizeMb` | number
`minimumChunkSizeMb` | number
`port` | number
`secretKey` | string
`serverSideAssembly` | boolean
`storagePath` | string

## Example

```typescript
import type { QuayBackendRGWConfiguration } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "accessKey": null,
  "bucketName": null,
  "hostname": null,
  "isSecure": null,
  "maximumChunkSizeMb": null,
  "minimumChunkSizeMb": null,
  "port": null,
  "secretKey": null,
  "serverSideAssembly": null,
  "storagePath": null,
} satisfies QuayBackendRGWConfiguration

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as QuayBackendRGWConfiguration
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


