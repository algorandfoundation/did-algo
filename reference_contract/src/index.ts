/* eslint-disable no-await-in-loop */
/* eslint-disable no-restricted-syntax */
import algosdk, { ABIMethod, ABIResult } from 'algosdk';
import { ApplicationClient } from '@algorandfoundation/algokit-utils/types/app-client';
import { AppCallTransactionResult } from '@algorandfoundation/algokit-utils/types/app';
import { expect } from '@jest/globals';
import { SuggestedParamsWithMinFee } from 'algosdk/dist/types/types/transactions/base';
import appSpec from '../contracts/artifacts/AlgoDID.json';

const COST_PER_BYTE = 400;
const COST_PER_BOX = 2500;
const MAX_BOX_SIZE = 32768;

const BYTES_PER_CALL = 2048
- 4 // 4 bytes for the method selector
- 34 // 34 bytes for the key
- 8 // 8 bytes for the box index
- 8; // 8 bytes for the offset

export type Metadata = {start: bigint, end: bigint, status: bigint, endSize: bigint};

export async function resolveDID(did: string, algodClient: algosdk.Algodv2): Promise<Buffer> {
  const splitDid = did.split(':');

  const idxOffset = splitDid.length === 6 ? 0 : 1;

  if (splitDid[0] !== 'did') {
    throw new Error(`invalid protocol, expected 'did', got ${splitDid[0]}`);
  }
  if (splitDid[1] !== 'algo') {
    throw new Error(`invalid DID method, expected 'algo', got ${splitDid[1]}`);
  }

  const nameSpace = splitDid[3 - idxOffset];

  if (nameSpace !== 'app') {
    throw new Error(`invalid namespace, expected 'app', got ${nameSpace}`);
  }

  const pubKey = Buffer.from(splitDid[5 - idxOffset], 'hex');

  let appID: bigint;

  try {
    appID = BigInt(splitDid[4 - idxOffset]);
    algosdk.encodeUint64(appID);
  } catch (e) {
    throw new Error(
      `invalid app ID, expected uint64, got ${splitDid[4 - idxOffset]}`,
    );
  }

  const appClient = new ApplicationClient({
    resolveBy: 'id',
    id: appID,
    sender: algosdk.generateAccount(),
    app: JSON.stringify(appSpec),
  }, algodClient);

  const boxValue = (await appClient.getBoxValueFromABIType(pubKey, algosdk.ABIType.from('(uint64,uint64,uint8,uint64,uint64)'))).valueOf() as bigint[];

  const metadata: Metadata = {
    start: boxValue[0], end: boxValue[1], status: boxValue[2], endSize: boxValue[3],
  };

  if (metadata.status === BigInt(0)) throw new Error('DID document is still being uploaded');
  if (metadata.status === BigInt(2)) throw new Error('DID document is being deleted');

  const boxPromises = [];
  for (let i = metadata.start; i <= metadata.end; i += 1n) {
    boxPromises.push(appClient.getBoxValue(algosdk.encodeUint64(i)));
  }

  const boxValues = await Promise.all(boxPromises);

  return Buffer.concat(boxValues);
}

/**
 *
 * @param algodClient
 * @param abiMethod
 * @param pubKey
 * @param boxes
 * @param boxIndex
 * @param suggestedParams
 * @param sender
 * @param appID
 * @param group
 * @returns
 */
export async function sendTxGroup(
  algodClient: algosdk.Algodv2,
  abiMethod: ABIMethod,
  bytesOffset: number,
  pubKey: Uint8Array,
  boxes: algosdk.BoxReference[],
  boxIndex: bigint,
  suggestedParams: SuggestedParamsWithMinFee,
  sender: algosdk.Account,
  appID: number,
  group: Buffer[],
): Promise<string[]> {
  const atc = new algosdk.AtomicTransactionComposer();
  group.forEach((chunk, i) => {
    atc.addMethodCall({
      method: abiMethod!,
      methodArgs: [pubKey, boxIndex, BYTES_PER_CALL * (i + bytesOffset), chunk],
      boxes,
      suggestedParams,
      sender: sender.addr,
      signer: algosdk.makeBasicAccountTransactionSigner(sender),
      appID,
    });
  });

  await new Promise((r) => setTimeout(r, 2000));
  return (await atc.execute(algodClient, 3)).txIDs;
}

/**
 *
 * @param atc
 * @param algodClient
 * @param retryCount
 */
async function tryExecute(
  atc: algosdk.AtomicTransactionComposer,
  algodClient: algosdk.Algodv2,
  retryCount = 1,
): Promise<void> {
  try {
    await atc.execute(algodClient, 3);
  } catch (e) {
    if (retryCount === 3) {
      // TODO: SDK bugfix
      // const execTraceConfig = new algosdk.modelsv2.SimulateTraceConfig({
      //   enable: true,
      //   stackChange: true,
      //   stateChange: true,
      //   scratchChange: true,
      // });

      // const simReq = new algosdk.modelsv2.SimulateRequest({
      //   txnGroups: [],
      //   execTraceConfig,
      // });
      // const result = await atc.simulate(algodClient, simReq);

      // console.warn(result.simulateResponse.txnGroups[0].txnResults[0].execTrace);
      throw e;
    }

    // eslint-disable-next-line no-console
    console.warn(`Failed to send transaction group. Retrying in ${500 * retryCount}ms (${retryCount / 3})`);
  }
}

/**
 *
 * @param data
 * @param appID
 * @param pubKey
 * @param sender
 * @param algodClient
 * @returns
 */
export async function uploadDIDDocument(
  data: Buffer,
  appID: number,
  pubKey: Uint8Array,
  sender: algosdk.Account,
  algodClient: algosdk.Algodv2,
): Promise<Metadata> {
  const appClient = new ApplicationClient({
    resolveBy: 'id',
    id: appID,
    sender,
    app: JSON.stringify(appSpec),
  }, algodClient);

  const ceilBoxes = Math.ceil(data.byteLength / MAX_BOX_SIZE);

  const endBoxSize = data.byteLength % MAX_BOX_SIZE;

  const totalCost = ceilBoxes * COST_PER_BOX // cost of data boxes
  + (ceilBoxes - 1) * MAX_BOX_SIZE * COST_PER_BYTE // cost of data
  + ceilBoxes * 8 * COST_PER_BYTE // cost of data keys
  + endBoxSize * COST_PER_BYTE // cost of last data box
  + COST_PER_BOX + (8 + 8 + 1 + 8 + 32 + 8) * COST_PER_BYTE; // cost of metadata box

  const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
    from: sender.addr,
    to: (await appClient.getAppReference()).appAddress,
    amount: totalCost,
    suggestedParams: await algodClient.getTransactionParams().do(),
  });

  const appCallResult: AppCallTransactionResult = await appClient.call({
    method: 'startUpload',
    methodArgs: [pubKey, ceilBoxes, endBoxSize, mbrPayment],
    boxes: [
      pubKey,
    ],
    sendParams: { suppressLog: true },
  });
  expect(appCallResult).toBeDefined();

  const boxValue = (await appClient.getBoxValueFromABIType(pubKey, algosdk.ABIType.from('(uint64,uint64,uint8,uint64,uint64)'))).valueOf() as bigint[];

  const metadata: Metadata = {
    start: boxValue[0], end: boxValue[1], status: boxValue[2], endSize: boxValue[3],
  };

  const numBoxes = Math.floor(data.byteLength / MAX_BOX_SIZE);
  const boxData: Buffer[] = [];

  for (let i = 0; i < numBoxes; i += 1) {
    const box = data.subarray(i * MAX_BOX_SIZE, (i + 1) * MAX_BOX_SIZE);
    boxData.push(box);
  }

  boxData.push(data.subarray(numBoxes * MAX_BOX_SIZE, data.byteLength));

  const suggestedParams: SuggestedParamsWithMinFee = await algodClient.getTransactionParams().do();

  const boxPromises = boxData.map(async (box, boxIndexOffset) => {
    const boxIndex = metadata.start + BigInt(boxIndexOffset);
    const numChunks = Math.ceil(box.byteLength / BYTES_PER_CALL);

    const chunks: Buffer[] = [];

    for (let i = 0; i < numChunks; i += 1) {
      chunks.push(box.subarray(i * BYTES_PER_CALL, (i + 1) * BYTES_PER_CALL));
    }

    const boxRef = { appIndex: 0, name: algosdk.encodeUint64(boxIndex) };
    const boxes: algosdk.BoxReference[] = new Array(7).fill(boxRef);

    boxes.push({ appIndex: 0, name: pubKey });

    const firstGroup = chunks.slice(0, 8);
    const secondGroup = chunks.slice(8);

    await sendTxGroup(algodClient, appClient.getABIMethod('upload')!, 0, pubKey, boxes, boxIndex, suggestedParams, sender, appID, firstGroup);

    if (secondGroup.length === 0) return;

    await sendTxGroup(algodClient, appClient.getABIMethod('upload')!, 8, pubKey, boxes, boxIndex, suggestedParams, sender, appID, secondGroup);
  });

  await Promise.all(boxPromises);
  if (Buffer.concat(boxData).toString('hex') !== data.toString('hex')) throw new Error('Data validation failed!');

  await appClient.call({
    method: 'finishUpload',
    methodArgs: [pubKey],
    boxes: [
      pubKey,
    ],
    sendParams: { suppressLog: true },
  });

  return metadata;
}

/*
export async function uploadDIDDocument(
  data: Buffer,
  appID: number,
  pubKey: Uint8Array,
  sender: algosdk.Account,
  algodClient: algosdk.Algodv2,
): Promise<Metadata> {
  */
export async function deleteDIDDocument(
  appID: number,
  pubKey: Uint8Array,
  sender: algosdk.Account,
  algodClient: algosdk.Algodv2,
): Promise<void> {
  const appClient = new ApplicationClient({
    resolveBy: 'id',
    id: appID,
    sender: algosdk.generateAccount(),
    app: JSON.stringify(appSpec),
  }, algodClient);

  const boxValue = (await appClient.getBoxValueFromABIType(pubKey, algosdk.ABIType.from('(uint64,uint64,uint8,uint64,uint64)'))).valueOf() as bigint[];

  const metadata: Metadata = {
    start: boxValue[0], end: boxValue[1], status: boxValue[2], endSize: boxValue[3],
  };

  await appClient.call({
    method: 'startDelete',
    methodArgs: [pubKey],
    boxes: [
      pubKey,
    ],
    sender,
    sendParams: { suppressLog: true },
  });

  const suggestedParams = await algodClient.getTransactionParams().do();

  const atcs: {boxIndex: bigint, atc: algosdk.AtomicTransactionComposer}[] = [];
  for (let boxIndex = metadata.start; boxIndex <= metadata.end; boxIndex += 1n) {
    const atc = new algosdk.AtomicTransactionComposer();
    const boxIndexRef = { appIndex: appID, name: algosdk.encodeUint64(boxIndex) };
    atc.addMethodCall({
      appID,
      method: appClient.getABIMethod('deleteData')!,
      methodArgs: [pubKey, boxIndex],
      boxes: [
        { appIndex: appID, name: pubKey },
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
      ],
      suggestedParams: { ...suggestedParams, fee: 2000, flatFee: true },
      sender: sender.addr,
      signer: algosdk.makeBasicAccountTransactionSigner(sender),
    });

    for (let i = 0; i < 4; i += 1) {
      atc.addMethodCall({
        appID,
        method: appClient.getABIMethod('dummy')!,
        methodArgs: [],
        boxes: [
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
        ],
        suggestedParams,
        sender: sender.addr,
        signer: algosdk.makeBasicAccountTransactionSigner(sender),
        note: new Uint8Array(Buffer.from(`dummy ${i}`)),
      });
    }

    atcs.push({ atc, boxIndex });
  }

  for await (const atcAndIndex of atcs) {
    await tryExecute(atcAndIndex.atc, algodClient);
  }
}

export async function updateDIDDocument(
  data: Buffer,
  appID: number,
  pubKey: Uint8Array,
  sender: algosdk.Account,
  algodClient: algosdk.Algodv2,
): Promise<Metadata> {
  await deleteDIDDocument(appID, pubKey, sender, algodClient);
  return uploadDIDDocument(data, appID, pubKey, sender, algodClient);
}
