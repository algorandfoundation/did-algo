# DID Method

[![Build Status](https://github.com/algorandfoundation/did-algo/workflows/ci/badge.svg?branch=main)](https://github.com/algorandfoundation/did-algo/actions)
[![Version](https://img.shields.io/github/tag/algorandfoundation/did-algo.svg)](https://github.com/algorandfoundation/did-algo/releases)
[![Software License](https://img.shields.io/badge/license-BSD3-red.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/algorandfoundation/did-algo?style=flat)](https://goreportcard.com/report/github.com/algorandfoundation/did-algo)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.0-ff69b4.svg)](.github/CODE_OF_CONDUCT.md)

The present document describes the __"algo"__ DID Method specification. The
definitions, conventions and technical details included intend to provide a
solid base for further developments while maintaining compliance with the work,
on the [W3C Credentials Community Group](https://w3c-ccg.github.io/did-spec/).

For more information about the origin and purpose of Decentralized Identifiers please
refer to the original [DID Primer.](https://github.com/WebOfTrustInfo/rwot5-boston/blob/master/topics-and-advance-readings/did-primer.md)

To facilitate adoption and testing, and promote open discussions about the subjects
treated, this repository also includes an open source reference implementation for a
CLI client and network agent. You can directly download a precompiled binary from the
[published releases](https://github.com/algorandfoundation/did-algo/releases).

## 1. Decentralized Identifiers

__In order to access online, i.e. digital, services, we need to be electronically
identifiable.__ It means we need an electronic profile that, with a certain level
of assurance, the service provider (either another person or an entity) can trust
it corresponds to our real identity.

__Conventional identity management systems are based on centralized authorities.__
These authorities establish a process by which to entitle a user temporary access
to a given identifier element. Nevertheless, the true ownership of the identifier
remains on the assigner side and thus, can be removed, revoked and reassigned if
deemed adequate. This creates and intrinsically asymmetric power relationship between
the authority entity and the user. Some examples of this kind of identifiers include:

- Domain names
- Email addresses
- Phone numbers
- IP addresses
- User names

Additionally, from the standpoint of cryptographic trust verification, each of these
centralized authorities serves as its own
[Root of Trust](https://csrc.nist.gov/Projects/Hardware-Roots-of-Trust).

__An alternative model to manage digital identifiers must be open and user-centric__.
It should be considered as such by satisfying at least the following considerations:

- Anyone must have access to freely register, publish and update as many identifiers
  as considered necessary.
- There should be no centralized authority required for the generation and assignment
  of identifiers.
- The end user must have true ownership of the assigned identifiers, i.e. no one but
  the user should be able to remove, revoke and/or reassign the user's identifiers.

This model is commonly referred to as __Decentralized Identifiers__, and allows us to
build a new __(3P)__ digital identity: __Private, Permanent__ and
__Portable__.

## 2. Access Considerations

In order to be considered open, the system must be publicly available. Any user
should be able to freely register, publish and update as many identifiers as
desired without the express authorization of any third party. This characteristic
of the model permits us to classify it as __censorship resistant.__

At the same time, this level of openness makes the model vulnerable to malicious
intentions and abuse. In such a way that a bad actor may prevent legitimate access
to the system by consuming the available resources. This kind of cyber-attack is
known as a [DoS (Denial-of-Service) attack](https://en.wikipedia.org/wiki/Denial-of-service_attack).

> In computing, a denial-of-service attack (DoS attack) is a cyber-attack in which
  the perpetrator seeks to make a machine or network resource unavailable to its
  intended users by temporarily or indefinitely disrupting services of a host
  connected to the Internet. Denial of service is typically accomplished by
  flooding the targeted machine or resource with superfluous requests in an attempt
  to overload systems and prevent some or all legitimate requests from being
  fulfilled.

The `algo` DID Method specification includes a __"Request Ticket"__ security
mechanism designed to mitigate risks of abuse while ensuring open access and
censorship resistance.

## 3. DID Method Specification

The method specification provides all the technical considerations, guidelines and
recommendations produced for the design and deployment of the DID method
implementation. The document is organized in 3 main sections.

1. __DID Schema.__ Definitions and conventions used to generate valid identifier
   instances.
2. __DID Document.__ Considerations on how to generate and use the DID document
   associated with a given identifier instance.
3. __Agent Protocol.__ Technical specifications detailing how to perform basic
  network operations, and the risk mitigation mechanisms in place, for tasks such as:
    - Publish a new identifier instance.
    - Update an existing identifier instance.
    - Resolve an existing identifier and retrieve the latest published version of
      its DID Document.

### 3.1 DID Schema

A Decentralized Identifier is defined as a [RFC3986](https://tools.ietf.org/html/rfc3986)
Uniform Resource Identifier, with a format based on the generic DID schema. Fore more
information you can refer to the
[original documentation](https://w3c.github.io/did-core/#did-syntax).

```abnf
did                = "did:" method-name ":" method-specific-id
method-name        = 1*method-char
method-char        = %x61-7A / DIGIT
method-specific-id = *( *idchar ":" ) 1*idchar
idchar             = ALPHA / DIGIT / "." / "-" / "_" / pct-encoded
pct-encoded        = "%" HEXDIG HEXDIG
```

Example of a simple Decentralized Identifier (DID).

```
did:example:123456789abcdefghi
```

Expanding on the previous definitions the `algo` DID Method specification use the
following format.

```abnf
did                = "did:algo:" [tag ":"] specific-idstring
tag                = 1*tagchar
specific-idstring  = depends on the particular use case
tagchar            = ALPHA / DIGIT / "." / "-"
```

The optional `tag` element provides a flexible namespace mechanism that can be used
to classify identifier instances into logical groups of arbitrary complexity.

The `specific-idstring` field does not impose any format requirements to ensure the
maximum level of flexibility to end users and implementers. The official implementation
however, proposes and recommends two formal modes for id strings.

A DID URL is a network location identifier for a specific resource. It can be used
to retrieve things like representations of DID subjects, verification methods, services,
specific parts of a DID document, or other resources.

The following is the ABNF definition using the syntax in [RFC5234]. It builds on the
basic DID Syntax. The path-abempty, query, and fragment components are defined in
[RFC3986]. All DID URLs MUST conform to the DID URL Syntax ABNF Rules.

```abnf
did-url = did path-abempty [ "?" query ] [ "#" fragment ]
```

#### 3.1.1 Mode UUID

The id string should be a randomly generated lower-case UUID v4 instance as defined by
[RFC4122](https://tools.ietf.org/html/rfc4122). The formal schema for the
`specific-idstring` field on this mode is the following.

```abnf
specific-idstring      = time-low "-" time-mid "-"
                         time-high-and-version "-"
                         clock-seq-and-reserved
                         clock-seq-low "-" node
time-low               = 4hexOctet
time-mid               = 2hexOctet
time-high-and-version  = 2hexOctet
clock-seq-and-reserved = hexOctet
clock-seq-low          = hexOctet
node                   = 6hexOctet
hexOctet               = hexDigit hexDigit
hexDigit               = "0" / "1" / "2" / "3" / "4" / "5" / "6" / "7" /
                         "8" / "9" / "a" / "b" / "c" / "d" / "e" / "f"
```

Example of a DID instance of mode UUID with a `tag` value of `c137`.

```abnf
did:algo:c137:02825c9d-6660-4f17-92db-2bd22c4ed902
```

#### 3.1.2 Mode Hash

The id string should be a randomly generated 32 bytes [SHA3-256](https://goo.gl/Wx8pTY)
hash value, encoded in hexadecimal format as a lower-case string of 64 characters.
The formal schema for the `specific-idstring` field on this mode is the following.

```abnf
specific-idstring = 32hexOctet
hexOctet          = hexDigit hexDigit
hexDigit          = "0" / "1" / "2" / "3" / "4" / "5" / "6" / "7" /
                    "8" / "9" / "a" / "b" / "c" / "d" / "e" / "f"
```

Example of a DID instance of mode hash with a `tag` value of `c137`.

```
did:algo:c137:85d48aebe67da2fdd273d03071de663d4fdd470cff2f5f3b8b41839f8b07075c
```

### 3.2 DID Document

A Decentralized Identifier, regardless of its particular method, can be resolved
to a standard resource describing the subject. This resource is called a
[DID Document](https://w3c.github.io/did-core/#dfn-did-documents), and typically
contains, among other relevant details, cryptographic material to support
authentication of the DID subject.

> A DID Document set of data describing the DID subject, including mechanisms, such
  as cryptographic public keys, that the DID subject or a DID delegate can use to
  authenticate itself and prove its association with the DID. A DID document might
  have one or more different representations.

The document is a Linked Data structure that ensures a high degree of flexibility
while facilitating the process of acquiring, parsing and using the contained
information. For the moment, the suggested encoding format for the document is
[JSON-LD](https://www.w3.org/TR/json-ld/). Other formats could be used in the future.

> The term Linked Data is used to describe a recommended best practice for exposing
  sharing, and connecting information on the Web using standards, such as URLs,
  to identify things and their properties. When information is presented as Linked
  Data, other related information can be easily discovered and new information can be
  easily linked to it. Linked Data is extensible in a decentralized way, greatly
  reducing barriers to large scale integration.

At the very least, the document must include the DID subject it's referring to under
the `id` key.

```json
{
  "@context": "https://www.w3.org/ns/did/v1",
  "id": "did:algo:c137:b616fca9-ad86-4be5-bc9c-0e3f8e27dc8d"
}
```

As it stands, this document is not very useful in itself. Other relevant details that
are often included in a DID Document are:

- [Created:](https://w3c-ccg.github.io/did-spec/#created-optional)
  Timestamp of the original creation.
- [Updated:](https://w3c-ccg.github.io/did-spec/#updated-optional)
  Timestamp of the most recent change.
- [Public Keys:](https://w3c-ccg.github.io/did-spec/#public-keys)
  Public keys are used for digital signatures, encryption and other cryptographic
  operations, which in turn are the basis for purposes such as authentication, secure
  communication, etc.
- [Authentication:](https://w3c-ccg.github.io/did-spec/#authentication)
  List the enabled mechanisms by which the DID subject can cryptographically prove
  that they are, in fact, associated with a DID Document.
- [Services:](https://w3c-ccg.github.io/did-spec/#service-endpoints)
  In addition to publication of authentication and authorization mechanisms, the
  other primary purpose of a DID Document is to enable discovery of service endpoints
  for the subject. A service endpoint may represent any type of service the subject
  wishes to advertise, including decentralized identity management services for
  further discovery, authentication, authorization, or interaction.

Additionally, the DID Document may include any other fields deemed relevant for the
particular use case or implementation.

Example of a more complete, and useful, DID Document.
```json
{
  "@context": [
    "https://www.w3.org/ns/did/v1",
    "https://w3id.org/security/v1"
  ],
  "id": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802",
  "created": "2021-10-15T02:13:56Z",
  "updated": "2021-10-19T21:14:53Z",
  "verificationMethod": [
    {
      "id": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802#master",
      "type": "Ed25519VerificationKey2018",
      "controller": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802",
      "publicKeyMultibase": "zCh9PDTZzeWxk2WdH4M1e8k2951D5D11jz7Uti9HRBGiK"
    }
  ],
  "authentication": [
    "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802#master"
  ],
  "service": [
    {
      "id": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802#algo-connect",
      "type": "did.algorand.foundation.ExternalService",
      "serviceEndpoint": "https://did.algorand.foundation",
      "extensions": [
        {
          "id": "algo-address",
          "version": "0.1.0",
          "data": [
            {
              "address": "Q4HSY6GM7AJGVSZGWPI5NZW2TJ4SIFHPSBXG4MCL72B5DAJL3PCCXIE3HI",
              "asset": "ALGO",
              "network": "testnet"
            }
          ]
        }
      ]
    }
  ]
}
```

Is important to note that the official specifications around service endpoints are
still in a very early stage at this point. Where appropriate or required the present
Method specification builds on it and introduces new considerations.

#### 3.2.1 Method Requirements

Building upon the base requirements and recommendations from the original
specification, the `algo` DID method introduces the following additional guidelines.

- The fields `created` and `updated` are required for all generated DID Documents.
- All service endpoints included in the DID Document may include an additional `data`
  field. Is recommended to include all extra parameters required for the particular
  service under this field.
- Supported verification methods and signature formats
  - [Ed25519](https://w3c-ccg.github.io/ld-cryptosuite-registry/#ed25519signature2018)
  - [RSA](https://w3c-ccg.github.io/ld-cryptosuite-registry/#rsasignature2018)
    (with a minimum length of 4096 bits).
  - [secp256k1](https://w3c-ccg.github.io/ld-cryptosuite-registry/#eddsasasignaturesecp256k1)

More information on the official keys and signatures formats is available at
[LD Cryptographic Suite Registry](https://w3c-ccg.github.io/ld-cryptosuite-registry/).

#### 3.2.2 Proofs

[proof:](https://www.w3.org/TR/vc-data-model/#proofs-signatures) Cryptographic proof
of the integrity of the DID Document according its subject. Recently it was removed
from the DID core document. This method still generates valid proofs for all mutations
performed on the DID documents and returns it under the `proof` element of all
resolved identifiers.

```json
{
  "document": "...",
  "proof": {
    "@context": [
      "https://w3id.org/security/v1"
    ],
    "type": "Ed25519Signature2018",
    "created": "2020-08-08T03:12:53Z",
    "domain": "did.algorandfoundation.org",
    "nonce": "3ec84acf8b301f3d7e0bba25a24b438a",
    "proofPurpose": "authentication",
    "verificationMethod": "did:algo:46389176-6109-4de7-bdb4-67e4fcf0230d#master",
    "proofValue": "QvVkJxTWHf6BQO5A/RzgqDoz6neKaagHWspwSeWqztWnjnt7Rlc73KKiHRs9++C2tdV3pZQtPiKDk6C7Q7nFAQ=="
  }
}
```

> More information about this change is [available here](https://github.com/w3c/did-core/issues/293).

### 3.3 Agent Protocol

The method implementation introduces the concept of a __network agent__. A network
agent is responsible for handling incoming client requests. It's very important to
note that the agent itself adheres to an operational protocol. The protocol is
independent of the data storage and message delivery mechanisms used. The method
protocol can be implemented using a __Distributed Ledger Platform__, as well as any
other infrastructure components suitable for the particular use case.

There are two main groups of operations available, __read__ and __write__. Write
operations are required when a user wishes to publish a new identifier record to
the network, or update the available information for an existing one. Read
operations enable resolution and retrieval of DID Documents and other relevant
assets published in the network.

#### 3.3.1 Request Ticket

As described earlier, a security mechanism is required to prevent malicious and
abusive activities. For these purposes, we introduce a __ticket__ requirement for all
write network operations. The mechanism is based on the original
[HashCash](http://www.hashcash.org/hashcash.pdf) algorithm and aims to mitigate
the following problems.

- __Discourage [DoS Attacks](https://en.wikipedia.org/wiki/Denial-of-service_attack)__.
  By making the user cover the “costs” of submitting a request for processing.
- __Prevent [Replay Attacks](https://en.wikipedia.org/wiki/Replay_attack)__.
  Validating the ticket was specifically generated for the request being processed.
- __Authentication__.
  Ensuring the user submitting the ticket is the owner of the DID, by incorporating
  a digital signature requirement that covers both the ticket details and the
  DID instance.

A request ticket has the following structure.

```protobuf
message Ticket {
  int64 timestamp = 1;
  int64 nonce_value = 2;
  string key_id = 3;
  bytes document = 4;
  bytes proof = 5;
  bytes signature = 6;
}
```

The client generates a ticket for the request using the following algorithm.

1. Let the __"bel"__ function be a method to produce a deterministic binary-encoded
   representation of a given input value using little endian byte order.
2. Let the __"hex"__ function be a method to produce a deterministic hexadecimal
   binary-encoded representation of a given input value.
3. __"timestamp"__ is set to the current UNIX time at the moment of creating the
   ticket.
4. __"nonce"__ is a randomly generated integer of 64bit precision.
5. __"key_id"__ is set to the identifier from the cryptographic key used to
   generate the ticket signature, MUST be enabled as an authentication key for the
   DID instance.
6. __"document"__ is set to the JSON-encoded DID Document to process.
7. __"proof"__ is set to the JSON-encoded valid proof for the DID Document to process.
8. A HashCash round is initiated for the ticket. The hash mechanism used MUST be
   SHA3-256 and the content submitted for each iteration of the round is a byte
   concatenation of the form:
   `"bel(timestamp) | bel(nonce) | hex(key_id) | document | proof"`.
9. The __"nonce"__ value of the ticket is atomically increased by one for each
   iteration of the round.
10. The ticket's __"challenge"__ is implicitly set to the produced hash from the
   HashCash round.
11. The __"signature"__ for the ticket is generated using the selected key of the DID
    and the obtained challenge value: `did.keys["key_id"].sign(challenge)`

Upon receiving a new write request the network agent validates the request ticket
using the following procedure.

1. Verify the ticket's `challenge` is valid by performing a HashCash
   verification.
2. Validate `document` are a properly encoded DID Document.
3. Validate `proof` is valid for the DID Document included in the ticket.
4. DID instance `method` value is properly set and supported by the agent.
5. Ensure `document` don’t include any private key. For security reasons no
   private keys should ever be published on the network.
6. Verify `signature` is valid.
    - For operations submitting a new entry, the key contents are obtained directly
      from the ticket contents. This ensures the user submitting the new DID instance
      is the one in control of the corresponding private key.
    - For operations updating an existing entry, the key contents are obtained from
      the previously stored record. This ensures the user submitting the request is
      the one in control of the original private key.
7. If the request is valid, the entry will be created or updated accordingly.

A sample implementation of the described __Request Ticket__ mechanism is available
[here](https://github.com/algorandfoundation/did-algo/blob/master/proto/v1/ticket.go).

#### 3.3.2 DID Resolution

The simplest mechanism to resolve a particular DID instance to the latest published
version of its corresponding DID Document is using the provided CLI client.

```shell
algoid get did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802
```

The resolution and data retrieval is done via the agent's HTTP interface, performing
an HTTP __GET__ request of the form:

`https://did.algorand.foundation/v1/retrieve/{{method}}/{{subject}}`

For example:

```bash
curl -v https://did.algorand.foundation/v1/retrieve/algo/85eda27f-eb7b-4a3b-9ea3-913188511802
```

If the subject is valid, and information has been published to the network, the
response will include the latest version available of its corresponding DID Document
encoded in JSON-LD with a __200__ status code. If no information is available the
response will be a JSON encoded error message with a __404__ status code.

```json
{
  "document": "...",
  "proof": "..."
}
```

You can also retrieve an existing subject using the provided SDK and RPC interface.
For example, using the Go client.

```go
// Error handling omitted for brevity
sub := "c137:eeb0c865-ce21-4ad6-baf8-5ba287ba8683"
response, _ := client.Retrieve(context.TODO(), proto.Request{Subject:sub})
if response.Ok {
  id := new(did.Identifier)
  id.Decode(response.Contents)
}
```

#### 3.3.3 DID Publishing and Update

To publish a new identifier instance or to update an existing one you can also use
the agent's HTTP interface or the provided SDK and clients.

When using HTTP the operation should be a __POST__ request with a properly
constructed and JSON-encoded request as the request's data. Binary data should be
encoded in standard [Base64](https://en.wikipedia.org/wiki/Base64) when transmitted
using JSON.

You can also publish and update a DID identifier instance using the provided SDK and
RPC interface. For example, using the Go client.

```go
// Error handling omitted for brevity
res, _ := client.Process(context.TODO(), request)
if res.Ok {
   // ...
}
```

## 4. Client Operations

> To enable the full functionality of DIDs and DID Documents on a particular
  distributed ledger or network (called the target system), a DID method
  specification MUST specify how each of the following CRUD operations is performed
  by a client. Each operation MUST be specified to the level of detail necessary to
  build and test interoperable client implementations with the target system.

The following sections provide detailed descriptions and examples of all required
CRUD base operations and some more advanced use cases. As described earlier, all
supported operations can be accessed using either the agent's HTTP interface or the
provided SDK and CLI client tool.

For brevity the following examples use the provided CLI client tool.

### 4.1 CRUD Operations

Basic operations enabling the users to create, read, update and delete identifier
instances.

#### 4.1.1 Create (Register)

To locally create a new DID instance.

```sh
algoid create [reference name]
```

The value provided for `reference name` is an easy-to-remember alias you choose for
the new identifier instance, __it won't have any use in the network context__.
The CLI also performs the following tasks for the newly generated identifier.

- Create a new `master` Ed25519 private key for the identifier
- Set the `master` key as an authentication mechanism for the identifier
- Generates a cryptographic integrity proof for the identifier using the `master` key

If required, the `master` key can be recovered using the selected `recovery-mode`,
for more information inspect the options available for the `create` command.

```
Creates a new DID locally

Usage:
  algoid register [flags]

Aliases:
  register, create, new

Examples:
algoid register [DID reference name]

Flags:
  -h, --help                    help for register
  -m, --method string           method value for the identifier instance (default "algo")
  -p, --passphrase              set a passphrase as recovery method for the primary key
  -s, --secret-sharing string   number of shares and threshold value: shares,threshold (default "3,2")
  -t, --tag string              tag value for the identifier instance
```

#### 4.1.2 Read (Verify)

You can retrieve a list of all your existing identifiers using the following command.

```sh
algoid list
```

The output produced will be something like this.

```
Name     DID
my-id    did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802
sample   did:algo:99dc4a30-7434-42e5-ac75-5f330be0ea0a
```

To inspect the DID Document of your local identifiers.

```
algoid info [reference name]
```

The generated document will be something similar for the following example.

```json
{
  "@context": [
    "https://www.w3.org/ns/did/v1",
    "https://w3id.org/security/v1"
  ],
  "id": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802",
  "created": "2021-10-15T02:13:56Z",
  "updated": "2021-10-19T21:14:53Z",
  "verificationMethod": [
    {
      "id": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802#master",
      "type": "Ed25519VerificationKey2018",
      "controller": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802",
      "publicKeyMultibase": "zCh9PDTZzeWxk2WdH4M1e8k2951D5D11jz7Uti9HRBGiK"
    }
  ],
  "authentication": [
    "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802#master"
  ],
  "service": [
    {
      "id": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802#algo-connect",
      "type": "did.algorand.foundation.ExternalService",
      "serviceEndpoint": "https://did.algorand.foundation",
      "extensions": [
        {
          "id": "algo-address",
          "version": "0.1.0",
          "data": [
            {
              "address": "Q4HSY6GM7AJGVSZGWPI5NZW2TJ4SIFHPSBXG4MCL72B5DAJL3PCCXIE3HI",
              "asset": "ALGO",
              "network": "testnet"
            }
          ]
        }
      ]
    }
  ]
}
```

#### 4.1.3 Update (Publish)

Whenever you wish to make one of your identifiers, in its current state, accessible
to the world, you can publish it to the network.

```sh
algoid sync sample
```

The CLI tool will generate the __Request Ticket__, submit the operation for
processing to the network and present the final result.

```
2021-10-19T17:12:39-05:00 DBG key selected for the operation: did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802#master
2021-10-19T17:12:39-05:00 INF publishing: my-id
2021-10-19T17:12:39-05:00 INF generating request ticket pow=8
2021-10-19T17:12:39-05:00 DBG ticket obtained: 0081d16b3603582c5f585eeaa432fe0aa1595b43b19ffaa2b829a92aac18148b
2021-10-19T17:12:39-05:00 DBG time: 1.673612ms (rounds completed 38)
2021-10-19T17:12:39-05:00 INF establishing connection to network agent: localhost:9090
2021-10-19T17:12:39-05:00 INF submitting request to the network
2021-10-19T17:12:42-05:00 DBG request status: true
2021-10-19T17:12:42-05:00 INF identifier: /ipfs/bafkreicvif7gb62oofiyposafedjfc6pmwqjr7qkfix7yu426zolmn3rxq
```

Once an identifier is published any user can retrieve and validate your DID document.
If you make local changes to your identifier, like adding a new cryptographic key or
service endpoint, and you wish these adjustments to be accessible to the rest of the
users, you'll need to publish it again.

### 4.2 DID Instance Management

The CLI client also facilitates some tasks required to manage a DID instance.

#### 4.2.1 Key Management

A DID Document list all public keys in use for the referenced DID instance. Public
keys are used for digital signatures, encryption and other cryptographic operations,
which in turn are the basis for purposes such as authentication, secure communication,
etc.

```
Manage cryptographic keys associated with the DID

Usage:
  algoid edit key [command]

Available Commands:
  add         Add a new cryptographic key for the DID
  remove      Remove an existing cryptographic key for the DID
```

To add a new cryptographic key to one of your identifiers you can use the `did key add`
command.

```
Add a new cryptographic key for the DID

Usage:
  algoid edit key add [flags]

Examples:
algoid edit key add [DID reference name] --name my-new-key --type ed --authentication

Flags:
  -a, --authentication   enable this key for authentication purposes
  -h, --help             help for add
  -n, --name string      name to be assigned to the newly added key (default "key-#")
  -t, --type string      type of cryptographic key: RSA (rsa), Ed25519 (ed) or secp256k1 (koblitz) (default "ed")
```

It will produce and properly add a public key entry. The cryptographic
integrity proof on the DID Document will also be updated accordingly.

```json
{
  "id": "did:algo:4d81bd52-2edb-4703-b8fc-b26d514a9c56#code-sign",
  "type": "Ed25519VerificationKey2018",
  "controller": "did:algo:4d81bd52-2edb-4703-b8fc-b26d514a9c56",
  "publicKeyMultibase": "zCh9PDTZzeWxk2WdH4M1e8k2951D5D11jz7Uti9HRBGiK"
}
```

You can also safely remove an existing key from your identifier using the
`edit key rm` command.

```
Remove an existing cryptographic key for the DID

Usage:
  algoid edit key remove [flags]

Aliases:
  remove, rm

Examples:
algoid edit key remove [DID reference name] [key name]
```

#### 4.2.2 Linked Data Signatures

The CLI client also facilitates the process of generating and validating [Linked Data
Signatures](https://w3c-dvcg.github.io/ld-signatures/).

```
Produce a linked digital proof document

Usage:
  algoid proof [flags]

Aliases:
  proof, sign

Examples:
algoid proof [DID reference name] --input "contents to sign"

Flags:
  -d, --domain string    domain value to use (default "did.algorand.foundation")
  -h, --help             help for proof
  -i, --input string     contents to sign
  -k, --key string       key to use to produce the proof (default "master")
  -p, --purpose string   specific intent for the proof (default "authentication")
```

For example, to create a new signature document from an existing file you can run the
following command.

```sh
cat file_to_sign | algoid sign my-id
```

The output produced will be a valid JSON-LD document containing the signature details.

```json
{
  "@context": [
    "https://w3id.org/security/v1"
  ],
  "type": "Ed25519Signature2018",
  "creator": "did:algo:4d81bd52-2edb-4703-b8fc-b26d514a9c56#master",
  "created": "2019-03-15T14:05:54Z",
  "domain": "did.algorandfoundation.org",
  "nonce": "f14d4619a39f7deb5a382bf32b220726",
  "signatureValue": "khqsBcnCViYm/3QFjgAQX2iOGDbNpsD5rPWsokWNLsBxhtRf79A+qV1f+9sphjVCxNP02jesOOni3t9zMCZbBw=="
}
```

You can save and share the produced JSON output. Other users will be able to verify the
integrity and authenticity of the signature using the `verify` command.

```sh
cat file_to_sign | algoid verify signature.json
```

The CLI will inspect the signature file, retrieve the DID Document for the creator
and use the public key to verify the integrity and authenticity of the signature.

```
2021-10-19T17:22:36-05:00 INF verifying proof document
2021-10-19T17:22:36-05:00 DBG load signature file
2021-10-19T17:22:36-05:00 DBG decoding contents
2021-10-19T17:22:36-05:00 DBG validating proof verification method
2021-10-19T17:22:38-05:00 INF proof is valid
```

#### 4.2.3 Service Management

As mentioned in earlier sections, one of the more relevant aspects of a DID Document
is its capability to list interaction mechanisms available for a particular subject.
This is done by including information of __Service Endpoints__ in the document. Using
the CLI client you can manage the services enabled for any of your identifiers.

```
Manage services enabled for the identifier

Usage:
  algoid edit service [command]

Available Commands:
  add         Register a new service entry for the DID
  remove      Remove an existing service entry for the DID
```

To add a new service you can use the `edit service add` command.

```
Register a new service entry for the DID

Usage:
  algoid edit service add [flags]

Examples:
algoid edit service add [DID] --name my-service --endpoint https://www.agency.com/user_id

Flags:
  -e, --endpoint string   main URL to access the service
  -h, --help              help for add
  -n, --name string       service's reference name (default "external-service-#")
  -t, --type string       type identifier for the service handler (default "did.algorand.foundation.ExternalService")
```

It will produce and properly add a service endpoint entry. The cryptographic
integrity proof on the DID Document will also be updated accordingly.

```json
{
  "id": "did:algo:85eda27f-eb7b-4a3b-9ea3-913188511802#algo-connect",
  "type": "did.algorand.foundation.ExternalService",
  "serviceEndpoint": "https://did.algorand.foundation",
}
```

You can also safely remove a service from your identifier using the
`edit service remove` command.

```sh
Remove an existing service entry for the DID

Usage:
  algoid edit service remove [flags]

Aliases:
  remove, rm

Examples:
algoid edit service remove [DID reference name] [service name]
```

## 5. Cryptography Notice

This distribution includes cryptographic software. The country in which you currently
reside may have restrictions on the import, possession, use, and/or re-export to another
country, of encryption software. BEFORE using any encryption software, please check your
country's laws, regulations and policies concerning the import, possession, or use, and
re-export of encryption software, to see if this is permitted.
See <http://www.wassenaar.org/> for more information.

The U.S. Government Department of Commerce, Bureau of Industry and Security (BIS), has
classified this software as Export Commodity Control Number (ECCN) 5D002.C.1, which
includes information security software using or performing cryptographic functions with
asymmetric algorithms. The form and manner of this distribution makes it eligible for
export under the License Exception ENC Technology Software Unrestricted (TSU) exception
(see the BIS Export Administration Regulations, Section 740.13) for both object code and
source code.
