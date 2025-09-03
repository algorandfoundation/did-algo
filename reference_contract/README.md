# Reference Contract

This is the reference implementation of the contract that will be used to store the DIDDocuments for `did:algo`. This is the contract that is used by the CLI tool by default when deploying a new contract.

It should be noted that the `did:algo` spec will work with any contract that implements the ABI interface defined in [the ARC4 JSON description](./contracts/artifacts/DIDAlgoStorage.arc4.json) and the `did:algo` [spec](../SPEC.md).

## Tests

Tests for the contract can be found at [here](./__test__/did-algo-storage.test.ts).

To run the tests, run `npm run test`. If changes to the contract have been made, be sure to build all artifacts with `npm run build`

### Testing a different implementation

If you wish to create your own implementation of a contract that implements the spec, you can test it using this test suite. To do so, update the app spec path in `src/index.ts`:

```ts
import appSpecJson from "../contracts/artifacts/DIDAlgoStorage.arc56.json";
```

And run `npm run test`.

If your contract is created with an ABI other than `createApplication()void` then you will also need to update the creation logic in `__test__/did-algo-storage.test.ts`

```ts
    const deployment = await factory.send.create({
      method: "createApplication",
    });
```

## Interacting With Contract

Golang code for interacting with this contract can be seen [here](../client/internal/main.go).

TypeScript code for interacting with this contract can be seen [here](./src/index.ts).
