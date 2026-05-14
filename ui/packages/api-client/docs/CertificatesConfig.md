
# CertificatesConfig


## Properties

Name | Type
------------ | -------------
`$schema` | string
`ironicHTTPSCertificate` | string
`ironicHTTPSKey` | string
`sslAPICertificateFullChain` | string
`sslAPICertificateKey` | string
`sslCACertificate` | string
`sslIngressCertificateFullChain` | string
`sslIngressCertificateKey` | string

## Example

```typescript
import type { CertificatesConfig } from '@enclave-wizard-ui/api-client'

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "ironicHTTPSCertificate": null,
  "ironicHTTPSKey": null,
  "sslAPICertificateFullChain": null,
  "sslAPICertificateKey": null,
  "sslCACertificate": null,
  "sslIngressCertificateFullChain": null,
  "sslIngressCertificateKey": null,
} satisfies CertificatesConfig

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as CertificatesConfig
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


