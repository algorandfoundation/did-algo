# Abstract

Decentralized Identifiers (DIDs) are a new type of identifier for that enable decentralized identification of various entities. DIDs are designed to enable individuals and organizations to generate their own identifiers using systems they trust. These new identifiers enable entities to prove control over them by authenticating using cryptographic proofs such as digital signatures.

Most DIDs involve a registry which is often, but not always, a distributed ledger technology (DLT). This specification utilize Algorand, a blockchain DLT, as the registry for identities. The initial method defined in this spec utilizes a stateful application on an Algorand network but this spec may be expanded in the future to allow further ways of resolving identity using Algorand.

# 1. Introduction

## 1.1 Goals

The primary goal of `did:algo` is to leverage the unique features of Algorand to provide a reliable decentralized resolution of DIDs. In particular, `did:algo` is designed to inherit the permission-less nature of Algorand and allow anyone to publish an identifier in the immutable ledger. Publishers of the DID and its corresponding DID document **MAY** be the subject that is identified, but `did:algo` also enables 3rd party entities to publish DID documents on behalf of others in a verifiable manner. For example, one user, identified by their ed25519 public key, may have multiple DIDs via the `did:algo` method provided by various entities.

# 2. Terminology

## Algorand app

A stateful smart contract that exists on the Algorand blockchain. Every app has a unique uint64 identifier.

## non-archival algod node

The Algorand node software that has the entirety of active state but not necessarily the full chain history

## box storage

A type of key-value storage available in Algorand apps

## Algorand address

The identifier for an account in the Algorand ledger. The address is derived from an ed25519 public key.

## auth address

The Algorand address that has authority to sign transactions for a given Algorand address

## ARC4 ABI

The Application Binary Interface (ABI) defined in ARC4 for Algorand smart contracts

# 3. The did:algo Format

The ABNF for the `did:algo` format is described below

```abnf
did-algo-format = "did:algo" [":" network] ":" namespace

network = "testnet" / "mainnet" / "betanet" / "custom" ; If omitted, the algorand-network is implicitly "mainnet"

namespace = app-namespace ; Currently only one namespace is supported, but there may be more in the future

app-namespace = "app:" algorand-app ":" hex-ed25519-key

algorand-app = 1*DIGIT ; The unsigned 64-bit integer for an application on the Algorand network
hex-ed25519-key = 64HEXDIG ; The public ed25519 key of the subject encoded in base16
```

# 4. Specification

## 4.1 Network

The `network` in the `did:algo` method is used to specify which Algorand network must be used to resolve the DID. To resolve DID, only a non-archival algod node is needed. Resolvers SHOULD check the genesis hash of the node they are using to resolve the DID and verify it matches the network in the DID. It should be noted that this is NOT a security measure. A malicious node could serve incorrect genesis hash or box data through the API so it is important for resolvers to use a trusted node for resolution.

The table of supported networks and their corresponding genesis hashes is below.

| Network | Genesis Hash (base64)                        |
| ------- | -------------------------------------------- |
| mainnet | wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8= |
| testnet | SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI= |
| betanet | mFgazF+2uRS1tMiL9dsj01hJGySEmPN28B/TjjvpVW0= |

If the network in the DID is `custom`, the genesis hash MUST NOT be verified by resolvers. The primary use-case for `custom` is development on a local network.

## 4.2 App Namespace

### 4.2.1 Controller

DIDDocuments created for `did:algo:app` DIDs SHOULD NOT have a controller field. Consumers of the DIDDocument MUST NOT assume the subject is the controller. In the reference implementation the controller is an Algorand address, which may have a dynamic auth address. This means ed25519 verification methods are not sufficient for verification and MUST NOT be used to identify the address that has permission to modify the in-app data.

In other app implementations there may be one or more entities that control the DIDDocument data. Since these entities do not have a DID, the only way to verify controllers for a `did:algo` DID is to inspect the TEAL code of the app.

If a controller field is set, it should be assumed that there are additional Algorand entities without a DID that may have control over the document.

### 4.2.2 Metadata Box

In order to support `did:algo:app` resolution, an application must contain a box with a 32 byte key corresponding to the `hex-ed25519-key` of the subject. The value of this box box must start with the ARC4 ABI tuple: `(uint64,uint64,uint8,uint64)` this data structure shall be referred to as the metadata box. Additional data MAY be in the metadata box, but it MUST NOT alter the four initial values defined here.

`metadata[0]` is a `uint64` indicating the key of the data box that contains the start of the DIDDocument for the subject.

`metadata[1]` is a `uint64` indicating the key of the data box that contains the end of the DIDdocument for the subject. Reading the contents of `metadata[0]` to `metadata[1]` (inclusive) sequentially will result in the full DIDDocument for the subject. If `metadata[0] == metadata[1]`, the entire DIDDocument is in a single box.

`metadata[2]` is a `uint8` containing the status of the binary data for the subject's DID document within the app. A value of `0` indicates the data is currently being uploaded and is not currently resolvable. A value of `1` indicates the data is ready and can be resolved. A value of `2` indicates the data is being deleted and is not resolvable.

`metadata[3]` is a `uint64` indicating the amount of bytes in the final box.

### 4.2.3 Data box

The data for the DIDDocument may be split across multiple boxes since each box can only hold 4 kilobytes of data. Data boxes are referenced via their `uint64` keys. DIDDocuments MUST be read and written sequentially across data boxes.
