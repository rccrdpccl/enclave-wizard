
# ClusterConfig


## Properties

Name | Type
------------ | -------------
`$schema` | string
`agentHosts` | [Array&lt;HostEntry&gt;](HostEntry.md)
`apiVIP` | string
`baseDomain` | string
`clusterName` | string
`defaultNtpServers` | Array&lt;string&gt;
`diskEncryption` | boolean
`ingressVIP` | string
`machineNetwork` | string
`masterMaxPods` | number
`pullSecret` | any
`rendezvousIP` | string
`sshPubPath` | string

## Example

```typescript
import type { ClusterConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "agentHosts": null,
  "apiVIP": null,
  "baseDomain": null,
  "clusterName": null,
  "defaultNtpServers": null,
  "diskEncryption": null,
  "ingressVIP": null,
  "machineNetwork": null,
  "masterMaxPods": null,
  "pullSecret": null,
  "rendezvousIP": null,
  "sshPubPath": null,
} satisfies ClusterConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ClusterConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


