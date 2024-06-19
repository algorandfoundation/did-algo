/* eslint-disable no-plusplus */
import * as algokit from '@algorandfoundation/algokit-utils';
import fs from 'fs';
import { ApplicationClient } from '@algorandfoundation/algokit-utils/types/app-client';
import {
  describe, expect, beforeAll, it, jest,
} from '@jest/globals';
import algosdk from 'algosdk';
import { algodClient, kmdClient } from './common';
import appSpec from '../contracts/artifacts/AlgoDID.json';
import {
  resolveDID, uploadDIDDocument, deleteDIDDocument, updateDIDDocument,
} from '../src/index';

jest.setTimeout(20000);

describe('Algorand DID', () => {
  /**
   * Large data (> 32k) to simulate a large DID Document
   * that needs to be put into multiple boxes
   */
  const bigData = fs.readFileSync(`${__dirname}/TEAL.pdf`);

  /**
   * Small data (< 32k) to simulate a small DID Document
   * that can fit into a single box
   */
  const smallJSONObject = { keyOne: 'foo', keyTwo: 'bar' };

  /** The public key for the user in the tests that has a big DID Document */
  const bigDataUserKey = algosdk.decodeAddress(algosdk.generateAccount().addr).publicKey;

  /** The public key for the user in the tests that has a small DID Document */
  const smallDataUserKey = algosdk.decodeAddress(algosdk.generateAccount().addr).publicKey;

  /** The public key for the user in the tests that updates their DID Document */
  const updateDataUserKey = algosdk.decodeAddress(algosdk.generateAccount().addr).publicKey;

  /** algokti appClient for interacting with the contract */
  let appClient: ApplicationClient;

  /** The account that will be used to create and call the contract */
  let sender: algosdk.Account;

  /** The ID of the contract */
  let appId: number;

  beforeAll(async () => {
    sender = await algokit.getDispenserAccount(algodClient, kmdClient);

    appClient = new ApplicationClient({
      resolveBy: 'id',
      id: 0,
      sender,
      app: JSON.stringify(appSpec),
    }, algodClient);

    await appClient.create({ method: 'createApplication', methodArgs: [], sendParams: { suppressLog: true } });

    await appClient.fundAppAccount({
      amount: algokit.microAlgos(100_000),
      sendParams: { suppressLog: true },
    });

    appId = Number((await appClient.getAppReference()).appId);
  });

  describe('uploadDIDDocument and Resolve', () => {
    it('(LARGE) DIDocument upload and resolve', async () => {
      const { appId } = await appClient.getAppReference();
      const pubkeyHex = Buffer.from(bigDataUserKey).toString('hex');

      // Large upload
      await uploadDIDDocument(bigData, Number(appId), bigDataUserKey, sender, algodClient);

      // Reconstruct DID from several boxes
      const resolvedData: Buffer = await resolveDID(`did:algo:custom:app:${appId}:${pubkeyHex}`, algodClient);
      expect(resolvedData.toString('hex')).toEqual(bigData.toString('hex'));
    });

    it('(SMALL) DIDocument upload and resolve', async () => {
      const { appId } = await appClient.getAppReference();
      const pubkeyHex = Buffer.from(smallDataUserKey).toString('hex');

      // Small upload
      await uploadDIDDocument(
        Buffer.from(JSON.stringify(smallJSONObject)),
        Number(appId),
        smallDataUserKey,
        sender,
        algodClient,
      );

      // Reconstruct DID from several boxes
      const resolvedData: Buffer = await resolveDID(`did:algo:custom:app:${appId}:${pubkeyHex}`, algodClient);
      expect(resolvedData.toString('hex')).toEqual(Buffer.from(JSON.stringify(smallJSONObject)).toString('hex'));
    });
  });

  describe('deleteDIDDocument', () => {
    const deleteDIDDocumentTest = async (userKey: Uint8Array) => {
      await deleteDIDDocument(appId, userKey, sender, algodClient);
      const pubkeyHex = Buffer.from(userKey).toString('hex');

      await expect(resolveDID(`did:algo:custom:app:${appId}:${pubkeyHex}`, algodClient)).rejects.toThrow();
    };

    it('deletes big (multi-box) data', async () => {
      await deleteDIDDocumentTest(bigDataUserKey);
    });

    it('deletes small (single-box) data', async () => {
      await deleteDIDDocumentTest(smallDataUserKey);
    });

    it('returns MBR', async () => {
      const { appAddress } = await appClient.getAppReference();
      const appAmount = (await algodClient.accountInformation(appAddress).do()).amount;

      expect(appAmount).toBe(100_000);
    });
  });

  describe('updateDocument', () => {
    beforeAll(async () => {
      // Initially upload the big data as the DID Document
      await uploadDIDDocument(
        bigData,
        appId,
        updateDataUserKey,
        sender,
        algodClient,
      );
    });

    it('uploads and resolves new data', async () => {
      // Update the DID Document to be the small data
      const data = Buffer.from(JSON.stringify(smallJSONObject));
      await updateDIDDocument(
        data,
        appId,
        updateDataUserKey,
        sender,
        algodClient,
      );

      const pubkeyHex = Buffer.from(updateDataUserKey).toString('hex');
      const resolvedData = await resolveDID(`did:algo:custom:app:${appId}:${pubkeyHex}`, algodClient);

      expect(resolvedData.toString()).toEqual(JSON.stringify(smallJSONObject));
    });
  });
});
