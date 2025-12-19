/* eslint-disable no-plusplus */
import * as algokit from "@algorandfoundation/algokit-utils";
import fs from "fs";
import { describe, expect, beforeAll, it, jest } from "@jest/globals";
import {
  resolveDID,
  uploadDIDDocument,
  deleteDIDDocument,
  updateDIDDocument,
  appSpec,
} from "../src/index";
import { Address } from "algosdk";
import { DidAlgoStorageClient, DidAlgoStorageFactory } from "../contracts/artifacts/DIDAlgoStorageClient";

jest.setTimeout(20000);

describe("Algorand DID", () => {
  const algorand = algokit.AlgorandClient.defaultLocalNet();
  /**
   * Large data (> 32k) to simulate a large DID Document
   * that needs to be put into multiple boxes
   */
  const bigData = Uint8Array.from(fs.readFileSync(`${__dirname}/TEAL.pdf`))

  /**
   * Small data (< 32k) to simulate a small DID Document
   * that can fit into a single box
   */
  const smallJSONObject = { keyOne: "foo", keyTwo: "bar" };

  /** The public key for the user in the tests that has a big DID Document */
  const bigDataUserKey = algorand.account.random().publicKey;

  /** The public key for the user in the tests that has a small DID Document */
  const smallDataUserKey = algorand.account.random().publicKey;

  /** The public key for the user in the tests that updates their DID Document */
  const updateDataUserKey = algorand.account.random().publicKey;

  /** algokti appClient for interacting with the contract */
  let appClient: DidAlgoStorageClient;

  let deployResult: any;

  /** The account that will be used to create and call the contract */
  let sender: Address;

  /** The ID of the contract */
  let appId: bigint;

  beforeAll(async () => {
    sender = await algorand.account.localNetDispenser();

    const factory = algorand.client.getTypedAppFactory(DidAlgoStorageFactory, {
        defaultSender: sender,
    });

    const result = await factory.deploy({ onUpdate: 'append', onSchemaBreak: 'append' });
    appClient = result.appClient;
    deployResult = result.result;

    // If app was just created fund the app account
    if (['create', 'replace'].includes(deployResult.operationPerformed)) {
        await algorand.send.payment({
        amount: (1).algo(),
        sender: sender,
        receiver: appClient.appAddress,
       });
    }
    appId = appClient.appId;
  });

  describe("uploadDIDDocument and Resolve", () => {
    it("(LARGE) DIDocument upload and resolve", async () => {
      const pubkeyHex = Buffer.from(bigDataUserKey).toString("hex");

      // Large upload
      await uploadDIDDocument(appClient,bigData, appId, bigDataUserKey, sender, algorand);

      // Reconstruct DID from several boxes
      const resolvedData: Buffer = await resolveDID(
        appClient,
        `did:algo:custom:app:${appId}:${pubkeyHex}`
      );
      expect(resolvedData.toString("hex")).toEqual(Buffer.from(bigData).toString("hex"));
    });

    it("(SMALL) DIDocument upload and resolve", async () => {
      const pubkeyHex = Buffer.from(smallDataUserKey).toString("hex");

      // Small upload
      await uploadDIDDocument(
        appClient,
        Uint8Array.from(JSON.stringify(smallJSONObject)),
        appId,
        smallDataUserKey,
        sender,
        algorand,
      );

      // Reconstruct DID from several boxes
      const resolvedData: Buffer = await resolveDID(
        appClient,
        `did:algo:custom:app:${appId}:${pubkeyHex}`
      );
      expect(resolvedData.toString("hex")).toEqual(
        Buffer.from(JSON.stringify(smallJSONObject)).toString("hex"),
      );
    });
  });

  describe("deleteDIDDocument", () => {
    const deleteDIDDocumentTest = async (userKey: Uint8Array) => {
      await deleteDIDDocument(appId, userKey, sender, algorand);
      const pubkeyHex = Buffer.from(userKey).toString("hex");

      await expect(
        resolveDID(appClient,`did:algo:custom:app:${appId}:${pubkeyHex}`),
      ).rejects.toThrow();
    };

    it("deletes big (multi-box) data", async () => {
      await deleteDIDDocumentTest(bigDataUserKey);
    });

    it("deletes small (single-box) data", async () => {
      await deleteDIDDocumentTest(smallDataUserKey);
    });

    it.skip("returns MBR", async () => {
      const appAmount = (
        await algorand.client.algod
          .accountInformation(appClient.appAddress)
          .do()
      ).amount;

      expect(appAmount).toBe(100_000n);
    });
  });

  describe("updateDocument", () => {
    beforeAll(async () => {
      // Initially upload the big data as the DID Document
      await uploadDIDDocument(
        appClient,
        bigData,
        appId,
        updateDataUserKey,
        sender,
        algorand,
      );
    });

    it("uploads and resolves new data", async () => {
      // Update the DID Document to be the small data
      const data = Uint8Array.from(JSON.stringify(smallJSONObject));
      await updateDIDDocument(appClient, data, appId, updateDataUserKey, sender, algorand);

      const pubkeyHex = Buffer.from(updateDataUserKey).toString("hex");
      const resolvedData = await resolveDID(
        appClient,
        `did:algo:custom:app:${appId}:${pubkeyHex}`,
      );

      expect(resolvedData.toString()).toEqual(JSON.stringify(smallJSONObject));
    });
  });
});
