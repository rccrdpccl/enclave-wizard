
# VASTVipPool


## Properties

Name | Type
------------ | -------------
`ipRanges` | [Array&lt;VASTIPRange&gt;](VASTIPRange.md)
`subnetCidr` | number

## Example

```typescript
import type { VASTVipPool } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "ipRanges": null,
  "subnetCidr": null,
} satisfies VASTVipPool

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as VASTVipPool
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


