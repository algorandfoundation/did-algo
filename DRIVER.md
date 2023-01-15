# DIF Universal Resolver Driver

The driver implementation resolves Decentralized Identifiers (DIDs) for the
`did:algo` method, based on the [W3C DID Core 1.0](https://www.w3.org/TR/did-core/)
and [DID Resolution](https://w3c-ccg.github.io/did-resolution/) specifications.

## Driver Interface

The driver can be invoked via HTTP GET requests of the form call to:

`<http://<resolver-url>/1.0/identifiers/did:algo:<did subject>`

The driver can, optionally, receive an `Accept` header that will affect the result
returned in the HTTP body and the `Content-Type` header.

If the `Accept` header provided is `application/ld+json;profile="https://w3id.org/did-resolution"`
the resolver with return a DID Resolution Result structure by default with the content type
`application/ld+json;profile="https://w3id.org/did-resolution";charset=utf-8`. This is also the
default behavior when no `Accept` header is provided.

Request:

```shell
curl -X GET <http://localhost:8080/1.0/identifiers/did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82>
```

Response:

```json
HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 1123
Content-Type: application/ld+json;profile="https://w3id.org/did-resolution";charset=utf-8
Date: Sun, 15 Jan 2023 17:39:30 GMT
Strict-Transport-Security: max-age=15724800; includeSubDomains
X-Content-Type-Options: nosniff
X-Resolver-Version: 0.4.0

{
    "@context": [
        "https://w3id.org/did-resolution/v1"
    ],
    "didDocument": {
        "@context": [
            "https://www.w3.org/ns/did/v1",
            "https://w3id.org/security/suites/ed25519-2020/v1",
            "https://w3id.org/security/suites/x25519-2020/v1"
        ],
        "authentication": [
            "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82#master",
            "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82#key-2"
        ],
        "id": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82",
        "verificationMethod": [
            {
                "controller": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82",
                "id": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82#master",
                "publicKeyMultibase": "zCh9PDTZzeWxk2WdH4M1e8k2951D5D11jz7Uti9HRBGiK",
                "type": "Ed25519VerificationKey2020"
            },
            {
                "controller": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82",
                "id": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82#key-2",
                "publicKeyMultibase": "zGrjYfS1jotUQPyNqib75XNpGsS4ZL1MoPaEVF19a4W5h",
                "type": "Ed25519VerificationKey2020"
            }
        ]
    },
    "didDocumentMetadata": {
        "created": "2022-03-23T19:15:10Z",
        "deactivated": false,
        "updated": "2022-03-23T20:11:30Z"
    },
    "didResolutionMetadata": {
        "contentType": "application/ld+json;profile=\"https://w3id.org/did-resolution\"",
        "retrieved": "2023-01-15T17:39:30Z"
    }
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
<http://localhost:8080/1.0/identifiers/did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82>
```

Response:

```json
HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 806
Content-Type: application/did+ld+json;charset=utf-8
Date: Sun, 15 Jan 2023 17:47:11 GMT
Strict-Transport-Security: max-age=15724800; includeSubDomains
X-Content-Type-Options: nosniff
X-Resolver-Version: 0.4.0

{
    "@context": [
        "https://www.w3.org/ns/did/v1",
        "https://w3id.org/security/suites/ed25519-2020/v1",
        "https://w3id.org/security/suites/x25519-2020/v1"
    ],
    "authentication": [
        "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82#master",
        "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82#key-2"
    ],
    "id": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82",
    "verificationMethod": [
        {
            "controller": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82",
            "id": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82#master",
            "publicKeyMultibase": "zCh9PDTZzeWxk2WdH4M1e8k2951D5D11jz7Uti9HRBGiK",
            "type": "Ed25519VerificationKey2020"
        },
        {
            "controller": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82",
            "id": "did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82#key-2",
            "publicKeyMultibase": "zGrjYfS1jotUQPyNqib75XNpGsS4ZL1MoPaEVF19a4W5h",
            "type": "Ed25519VerificationKey2020"
        }
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
<http://localhost:8080/1.0/identifiers/did:algo:c93fdef1-8a0a-4c65-8c54-fd33117c9e82>
```

Response:

```json
HTTP/1.1 406 Not Acceptable
Connection: keep-alive
Content-Length: 187
Content-Type: application/ld+json;profile="https://w3id.org/did-resolution";charset=utf-8
Date: Sun, 15 Jan 2023 17:49:49 GMT
Strict-Transport-Security: max-age=15724800; includeSubDomains
X-Content-Type-Options: nosniff
X-Resolver-Version: 0.4.0

{
    "@context": [
        "https://w3id.org/did-resolution/v1"
    ],
    "didResolutionMetadata": {
        "contentType": "application/did+cbor",
        "error": "representationNotSupported",
        "retrieved": "2023-01-15T17:49:49Z"
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
