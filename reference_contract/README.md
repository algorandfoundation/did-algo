# Reference Contract

This is the reference implementation of the contract that will be used to store the DIDDocuments for `did:algo`. This is the contract that is used by the CLI tool by default when deploying a new contract.

It should be noted that the `did:algo` spec will work with any contract that implements the ABI interface defined in [the ARC4 JSON description](./contracts/artifacts/DIDAlgoStorage.arc4.json) and the `did:algo` [spec](../SPEC.md).

## Tests

Tests for the contract can be found at [here](./__test__/did-algo-storage.test.ts).

## Interacting With Contract

Golang code for interacting with this contract can be seen [here](../client/internal/main.go).

TypeScript code for interacting with this contract can be seen [here](./src/index.ts).
