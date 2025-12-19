/* eslint-disable no-await-in-loop */
/* eslint-disable no-restricted-syntax */
import algosdk, { ABIMethod, ABIResult, Address } from "algosdk";
import { expect } from "@jest/globals";
import appSpecJson from "../contracts/artifacts/DIDAlgoStorage.arc56.json";
import { AppClient } from "@algorandfoundation/algokit-utils/types/app-client";
import { AlgorandClient, microAlgos } from "@algorandfoundation/algokit-utils";
import { BoxReference } from "@algorandfoundation/algokit-utils/types/app-manager";
import { TransactionComposer } from "@algorandfoundation/algokit-utils/types/composer";
import { DidAlgoStorageClient } from "../contracts/artifacts/DIDAlgoStorageClient";

export const appSpec = JSON.stringify(appSpecJson);

const COST_PER_BYTE = 400;
const COST_PER_BOX = 2500;
const MAX_BOX_SIZE = 32768;

const BYTES_PER_CALL =
  2048 -
  4 - // 4 bytes for the method selector
  34 - // 34 bytes for the key
  8 - // 8 bytes for the box index
  8; // 8 bytes for the offset

export type Metadata = {
  start: bigint;
  end: bigint;
  status: bigint;
  endSize: bigint;
};

export async function resolveDID(
  appClient: DidAlgoStorageClient,
  did: string
): Promise<Buffer> {
  const splitDid = did.split(":");

  const idxOffset = splitDid.length === 6 ? 0 : 1;

  if (splitDid[0] !== "did") {
    throw new Error(`invalid protocol, expected 'did', got ${splitDid[0]}`);
  }
  if (splitDid[1] !== "algo") {
    throw new Error(`invalid DID method, expected 'algo', got ${splitDid[1]}`);
  }

  const nameSpace = splitDid[3 - idxOffset];

  if (nameSpace !== "app") {
    throw new Error(`invalid namespace, expected 'app', got ${nameSpace}`);
  }

  const pubKey = new Uint8Array(Buffer.from(splitDid[5 - idxOffset], "hex"));

  let appID: bigint;

  try {
    appID = BigInt(splitDid[4 - idxOffset]);
    algosdk.encodeUint64(appID);
  } catch (e) {
    throw new Error(
      `invalid app ID, expected uint64, got ${splitDid[4 - idxOffset]}`,
    );
  }

  // const boxValue = (
  //   await appClient.getBoxValueFromABIType(
  //     pubKey,
  //     algosdk.ABIType.from("(uint64,uint64,uint8,uint64,uint64)"),
  //   )
  // ).valueOf() as bigint[];
  const boxValue: any = await appClient.state.box.metadata.value(algosdk.encodeAddress(pubKey));


  const metadata: Metadata = {
    start: boxValue["start"],
    end: boxValue["end"],
    status: boxValue["status"],
    endSize: boxValue["endSize"],
  };

  if (metadata.status === BigInt(0))
    throw new Error("DID document is still being uploaded");
  if (metadata.status === BigInt(2))
    throw new Error("DID document is being deleted");

  const boxPromises = [];
  for (let i = metadata.start; i <= metadata.end; i += 1n) {
    boxPromises.push(appClient.state.box.dataBoxes.value(i));
  }

  const boxValues = await Promise.all(boxPromises);

  return Buffer.concat(boxValues.filter((v): v is Uint8Array => v !== undefined));
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
  algorand: AlgorandClient,
  abiMethod: ABIMethod,
  bytesOffset: number,
  pubKey: Uint8Array,
  boxes: BoxReference[],
  boxIndex: bigint,
  sender: Address,
  appID: bigint,
  group: Uint8Array[],
): Promise<string[]> {
  const composer = algorand.newGroup();
  group.forEach((chunk, i) => {

    composer.addAppCallMethodCall({
      method: abiMethod!,
      args: [pubKey, boxIndex, BYTES_PER_CALL * (i + bytesOffset), chunk],
      boxReferences: boxes,
      sender,
      appId: appID,
    });
  });

  await new Promise((r) => setTimeout(r, 2000));
  return (await composer.send({ maxRoundsToWaitForConfirmation: 3 })).txIds;
}

/**
 *
 * @param atc
 * @param algodClient
 * @param retryCount
 */
async function tryExecute(
  composer: TransactionComposer,
  retryCount = 1,
): Promise<void> {
  try {
    await composer.send({ maxRoundsToWaitForConfirmation: 3 });
  } catch (e) {
    if (retryCount === 3) {
      throw e;
    }

    // eslint-disable-next-line no-console
    console.warn(
      `Failed to send transaction group. Retrying in ${500 * retryCount}ms (${retryCount / 3})`,
    );
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
  appClient: DidAlgoStorageClient,
  data: Buffer,
  appID: bigint,
  pubKey: Uint8Array,
  sender: Address,
  algorand: AlgorandClient,
): Promise<Metadata> {

  const ceilBoxes = Math.ceil(data.byteLength / MAX_BOX_SIZE);
  const endBoxSize = data.byteLength % MAX_BOX_SIZE;

  const totalCost =
    ceilBoxes * COST_PER_BOX + // cost of data boxes
    (ceilBoxes - 1) * MAX_BOX_SIZE * COST_PER_BYTE + // cost of data
    ceilBoxes * 8 * COST_PER_BYTE + // cost of data keys
    endBoxSize * COST_PER_BYTE + // cost of last data box
    COST_PER_BOX +
    (8 + 8 + 1 + 8 + 32 + 8) * COST_PER_BYTE; // cost of metadata box

  const mbrPayment = appClient.algorand.createTransaction.payment({
    sender,
    receiver: appClient.appAddress,
    amount: microAlgos(totalCost),
  });

  const appCallResult = await appClient.send.startUpload({
    args: [algosdk.encodeAddress(pubKey), ceilBoxes, endBoxSize, mbrPayment],
    boxReferences: [pubKey],
  });
  expect(appCallResult).toBeDefined();

  // const boxValue = (
  //   await appClient.getBoxValueFromABIType(
  //     pubKey,
  //     algosdk.ABIType.from("(uint64,uint64,uint8,uint64,uint64)"),
  //   )
  // ).valueOf() as bigint[];

  // Using ApplicationClient / AppClient state helpers
  
  const boxValue: any = await appClient.state.box.metadata.value(algosdk.encodeAddress(pubKey));

  const metadata: Metadata = {
    start: boxValue["start"],
    end: boxValue["end"],
    status: boxValue["status"],
    endSize: boxValue["endSize"],
  };

  const numBoxes = Math.floor(data.byteLength / MAX_BOX_SIZE);
  const boxData: Uint8Array[] = [];

  for (let i = 0; i < numBoxes; i += 1) {
    const box = new Uint8Array(data.subarray(i * MAX_BOX_SIZE, (i + 1) * MAX_BOX_SIZE));
    boxData.push(box);
  }

  boxData.push(new Uint8Array(data.subarray(numBoxes * MAX_BOX_SIZE, data.byteLength)));
  const boxPromises = boxData.map(async (box, boxIndexOffset) => {
    const boxIndex = metadata.start + BigInt(boxIndexOffset);
    const numChunks = Math.ceil(box.byteLength / BYTES_PER_CALL);

    const chunks: Uint8Array[] = [];

    for (let i = 0; i < numChunks; i += 1) {
      chunks.push(box.subarray(i * BYTES_PER_CALL, (i + 1) * BYTES_PER_CALL));
    }
    const boxRef: BoxReference = {
      appId: 0n, // 0 == means self
      name: algosdk.encodeUint64(boxIndex),
    };

    // 7 refs data + 1 metadata
    const boxes: BoxReference[] = new Array(7).fill(boxRef);
    boxes.push({ appId: 0n, name: pubKey });

   const firstGroup = chunks.slice(0, 8);
   const secondGroup = chunks.slice(8);

   await sendTxGroup(
     algorand,
     appClient.appClient.getABIMethod("upload")!,
     0,
     pubKey,
     boxes,
     boxIndex,
     sender,
     appID,
     firstGroup,
   );

   // In case we need to write the full 32kb of the group, since we need some space for args and can only write 2k per call;
   // we will need a second group, to fully write the box with 32kb of data
   if (secondGroup.length === 0) return;

   await sendTxGroup(
     algorand,
     appClient.appClient.getABIMethod("upload")!,
     8,
     pubKey,
     boxes,
     boxIndex,
     sender,
     appID,
     secondGroup,
   );
  });

  await Promise.all(boxPromises);
  if (Buffer.concat(boxData).toString("hex") !== data.toString("hex"))
    throw new Error("Data validation failed!");

  await appClient.send.finishUpload({
    args: [algosdk.encodeAddress(pubKey)],
    boxReferences: [pubKey],
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
  appID: bigint,
  pubKey: Uint8Array,
  sender: Address,
  algorand: AlgorandClient,
): Promise<void> {
  const appClient = new AppClient({
    appId: appID,
    defaultSender: algorand.account.random(),
    appSpec,
    algorand,
  });

  const boxValue = (
    await appClient.getBoxValueFromABIType(
      pubKey,
      algosdk.ABIType.from("(uint64,uint64,uint8,uint64,uint64)"),
    )
  ).valueOf() as bigint[];

  const metadata: Metadata = {
    start: boxValue[0],
    end: boxValue[1],
    status: boxValue[2],
    endSize: boxValue[3],
  };

  await appClient.send.call({
    method: "startDelete",
    args: [pubKey],
    boxReferences: [pubKey],
    sender,
  });

  const composers: {
    boxIndex: bigint;
    composer: TransactionComposer;
  }[] = [];
  for (
    let boxIndex = metadata.start;
    boxIndex <= metadata.end;
    boxIndex += 1n
  ) {
    const composer = algorand.newGroup();
    const boxIndexRef = {
      appId: appID,
      name: algosdk.encodeUint64(boxIndex),
    };
    composer.addAppCallMethodCall({
      appId: appID,
      method: appClient.getABIMethod("deleteData")!,
      args: [pubKey, boxIndex],
      boxReferences: [
        { appId: appID, name: pubKey },
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
        boxIndexRef,
      ],
      extraFee: microAlgos(1000),
      sender,
    });

    for (let i = 0; i < 4; i += 1) {
      composer.addAppCallMethodCall({
        appId: appID,
        method: appClient.getABIMethod("dummy")!,
        args: [],
        boxReferences: [
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
          boxIndexRef,
        ],
        sender,
        note: new Uint8Array(Buffer.from(`dummy ${i}`)),
      });
    }

    composers.push({ composer, boxIndex });
  }

  for await (const composerAndIndex of composers) {
    await tryExecute(composerAndIndex.composer);
  }
}

export async function updateDIDDocument(
  appClient: DidAlgoStorageClient,
  data: Buffer,
  appID: bigint,
  pubKey: Uint8Array,
  sender: Address,
  algorand: AlgorandClient,
): Promise<Metadata> {
  await deleteDIDDocument(appID, pubKey, sender, algorand);
  return uploadDIDDocument(appClient, data, appID, pubKey, sender, algorand);
}
