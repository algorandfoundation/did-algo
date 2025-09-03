/* eslint-disable no-plusplus */
import * as algokit from "@algorandfoundation/algokit-utils";
import fs from "fs";
import { describe, expect, beforeAll, it, jest } from "@jest/globals";
import appSpec from "../contracts/artifacts/DIDAlgoStorage.arc56.json";
import {
  resolveDID,
  uploadDIDDocument,
  deleteDIDDocument,
  updateDIDDocument,
} from "../src/index";
import { AppClient } from "@algorandfoundation/algokit-utils/types/app-client";
import { AppFactory } from "@algorandfoundation/algokit-utils/types/app-factory";
import { Address } from "algosdk";

jest.setTimeout(20000);

describe("Algorand DID", () => {
  const algorand = algokit.AlgorandClient.defaultLocalNet();
  /**
   * Large data (> 32k) to simulate a large DID Document
   * that needs to be put into multiple boxes
   */
  const bigData = fs.readFileSync(`${__dirname}/TEAL.pdf`);

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
  let appClient: AppClient;

  /** The account that will be used to create and call the contract */
  let sender: Address;

  /** The ID of the contract */
  let appId: bigint;

  beforeAll(async () => {
    sender = await algorand.account.localNetDispenser();

    const factory = new AppFactory({
      algorand,
      defaultSender: sender,
      appSpec: JSON.stringify(appSpec),
    });

    const deployment = await factory.send.create({
      method: "createApplication",
    });

    appClient = deployment.appClient;

    await appClient.fundAppAccount({
      amount: algokit.microAlgos(100_000),
    });

    appId = appClient.appId;
  });

  describe("uploadDIDDocument and Resolve", () => {
    it("(LARGE) DIDocument upload and resolve", async () => {
      const pubkeyHex = Buffer.from(bigDataUserKey).toString("hex");

      // Large upload
      await uploadDIDDocument(bigData, appId, bigDataUserKey, sender, algorand);

      // Reconstruct DID from several boxes
      const resolvedData: Buffer = await resolveDID(
        `did:algo:custom:app:${appId}:${pubkeyHex}`,
        algorand,
      );
      expect(resolvedData.toString("hex")).toEqual(bigData.toString("hex"));
    });

    it("(SMALL) DIDocument upload and resolve", async () => {
      const pubkeyHex = Buffer.from(smallDataUserKey).toString("hex");

      // Small upload
      await uploadDIDDocument(
        Buffer.from(JSON.stringify(smallJSONObject)),
        appId,
        smallDataUserKey,
        sender,
        algorand,
      );

      // Reconstruct DID from several boxes
      const resolvedData: Buffer = await resolveDID(
        `did:algo:custom:app:${appId}:${pubkeyHex}`,
        algorand,
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
        resolveDID(`did:algo:custom:app:${appId}:${pubkeyHex}`, algorand),
      ).rejects.toThrow();
    };

    it("deletes big (multi-box) data", async () => {
      await deleteDIDDocumentTest(bigDataUserKey);
    });

    it("deletes small (single-box) data", async () => {
      await deleteDIDDocumentTest(smallDataUserKey);
    });

    it("returns MBR", async () => {
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
        bigData,
        appId,
        updateDataUserKey,
        sender,
        algorand,
      );
    });

    it("uploads and resolves new data", async () => {
      // Update the DID Document to be the small data
      const data = Buffer.from(JSON.stringify(smallJSONObject));
      await updateDIDDocument(data, appId, updateDataUserKey, sender, algorand);

      const pubkeyHex = Buffer.from(updateDataUserKey).toString("hex");
      const resolvedData = await resolveDID(
        `did:algo:custom:app:${appId}:${pubkeyHex}`,
        algorand,
      );

      expect(resolvedData.toString()).toEqual(JSON.stringify(smallJSONObject));
    });
  });
});
