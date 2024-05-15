# DIF Universal Resolver Driver

The driver implementation resolves Decentralized Identifiers (DIDs) for the
`did:algo` method, based on the [W3C DID Core 1.0](https://www.w3.org/TR/did-core/)
and [DID Resolution](https://w3c-ccg.github.io/did-resolution/) specifications.

## Driver Interface

The driver can be invoked via HTTP GET requests of the form:

`<http://<resolver-url>/1.0/identifiers/did:algo:<did subject>`

The driver can, optionally, receive an `Accept` header that will affect the result
returned in the HTTP body and the `Content-Type` header.

If the `Accept` header provided is `application/ld+json;profile="https://w3id.org/did-resolution"`
the resolver with return a DID Resolution Result structure by default with the content type
`application/ld+json;profile="https://w3id.org/did-resolution";charset=utf-8`. This is also the
default behavior when no `Accept` header is provided.

Request:

```shell
curl -X GET <http://localhost:8080/1.0/identifiers/did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada>
```

Response:

```json
{
  "@context": [
    "https://www.w3.org/ns/did/v1",
    "https://w3id.org/security/suites/ed25519-2020/v1",
    "https://w3id.org/security/suites/x25519-2020/v1"
  ],
  "id": "did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada",
  "verificationMethod": [
    {
      "id": "did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada#master",
      "type": "Ed25519VerificationKey2020",
      "controller": "did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada",
      "publicKeyMultibase": "zFh6VYTmcxGD2vTNMSTHhtMSa7TkGCded9ofBX5C6CxYq"
    }
  ],
  "authentication": [
    "did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada#master"
  ]
}
```

The following `Accept` values will instruct the resolver to return a DID Document with
`Content-Type` set as `application/did+ld+json;charset=utf-8`.

- `application/json`
- `application/ld+json`
- `application/did+ld+json`

Request:

```shell
curl -X GET \
--header "Accept: application/did+ld+json" \
<http://localhost:8080/1.0/identifiers/did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada>
```

Response:

```json
{
  "@context": [
    "https://www.w3.org/ns/did/v1",
    "https://w3id.org/security/suites/ed25519-2020/v1",
    "https://w3id.org/security/suites/x25519-2020/v1"
  ],
  "id": "did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada",
  "verificationMethod": [
    {
      "id": "did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada#master",
      "type": "Ed25519VerificationKey2020",
      "controller": "did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada",
      "publicKeyMultibase": "zFh6VYTmcxGD2vTNMSTHhtMSa7TkGCded9ofBX5C6CxYq"
    }
  ],
  "authentication": [
    "did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada#master"
  ]
}
```

## Custom Representations

All other `Accept` values will instruct the resolver to process a "Resolve Representation"
request. If the mime type corresponds to an encoder registered in the resolver instance, it
will be used to generate the representation and return it in the response body.

If the mime type requested is not available in the resolver instance a `representationNotSupported`
error will be returned.

Request:

```shell
curl -X GET \
--header "Accept: application/did+cbor" \
<http://localhost:8080/1.0/identifiers/did:algo:mainnet:app:1845671812:da490f2d15a625459bf970a3d55e1a646dfd3a956d011546e953e945d39fdada>
```

Response:

```json
{
  "@context": ["https://w3id.org/did-resolution/v1"],
  "didResolutionMetadata": {
    "contentType": "application/did+cbor",
    "retrieved": "2024-03-01T01:39:03Z",
    "error": "representationNotSupported"
  }
}
```

Custom encoders can be provided by 3rd parties by compling with the interface.

```go
type Encoder interface {
  // Encode an existing DID document to a valid representation.
  // If an error is returned is must be a valid error code as
  // defined in the spec.
  Encode(doc *did.Document) ([]byte, error)
}
```

More information, and resolver package source code, is available here
<https://pkg.go.dev/go.bryk.io/pkg/did/resolver#pkg-overview>.
