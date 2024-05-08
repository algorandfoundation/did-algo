# Abstract

Decentralized Identifiers (DIDs) are a new type of identifier for that enable decentralized dentification of various entities. DIDs are designed to enable individuals and organizations to generate their own identifiers using systems they trust. These new identifiers enable entities to prove control over them by authenticating using cryptographic proofs such as digital signatures.

Most DIDs involve a registry which is often, but not always, a distributed ledger technology (DLT). This specification utilize Algorand, a blockchain DLT, as the registry for identities. The initial method defined in this spec utilizes a stateful application on an Algorand network but this spec may be expanded in the future to allow further ways of resolving identity using Algorand.

# 1. Introduction

## 1.1 Goals

The primary goal of `did:algo` is to leverage the unique features of Algorand to provide a reliable decentralized resolution of DIDs. In paticular, `did:algo` is designed to inherit the permisionless nature of Algorand and allow anyone to publish an identifier in the immutable ledger. Publishers of the DID and its corresponding DID document **MAY** be the subject that is identified, but `did:algo` also enables 3rd party entites to publish DID documents on behalf of others in a verifiable manner. For example, one user, identified by their ed25519 public key, may have multiple DIDs via the `did:algo` method provided by various entities.

# 2. Terminology

## Algorand app

A stateful smart contract that exists on the Algorand blockchain. Every app has a unique uint64 identifier.

## non-archival algod node

The Algorand node software that has the entirity of active state but not necessarily the full chain history

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
did-algo-format = "did:algo" [":" algorand-network] ":" algorand-namespace ":" namespace-format

algorand-network = "testnet" / "mainnet" ; If omitted, the algorand-network is implicity "mainnet"
algorand-namepspace = app-namespace

app-namespace = algorand-app ":" hex-ed25519key

algorand-app = 1*DIGIT ; The unsigned 64-bit integer for an application on the Algorand network
hex-ed25519key = 64HEXDIG ; The public ed25519 key of the subject encoded in base16
```

# 4. Implementation

TODO
