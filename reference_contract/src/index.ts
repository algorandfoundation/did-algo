
import algosdk, { ABIMethod, Address } from "algosdk";
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
const BYTES_PER_CALL = 2048 - 4 - 34 - 8 - 8;

export type Metadata = {
  start: bigint;
  end: bigint;
  status: bigint;
  endSize: bigint;
};

export async function resolveDID(appClient: DidAlgoStorageClient, did: string): Promise<Buffer> {
  const split = did.split(":");
  const idx = split.length === 6 ? 0 : 1;
  if (split[0] !== "did" || split[1] !== "algo" || split[3 - idx] !== "app") {
    throw new Error("Invalid DID format");
  }
  const pubKey = new Uint8Array(Buffer.from(split[5 - idx], "hex"));
  let appID: bigint;
  try {
    appID = BigInt(split[4 - idx]);
    algosdk.encodeUint64(appID);
  } catch {
    throw new Error("Invalid app ID");
  }
  const boxValue: any = await appClient.state.box.metadata.value(algosdk.encodeAddress(pubKey));
  const metadata: Metadata = {
    start: boxValue.start,
    end: boxValue.end,
    status: boxValue.status,
    endSize: boxValue.endSize,
  };
  if (metadata.status === 0n) throw new Error("DID document is still being uploaded");
  if (metadata.status === 2n) throw new Error("DID document is being deleted");
  const boxValues = await Promise.all(
    Array.from({ length: Number(metadata.end - metadata.start + 1n) }, (_, i) =>
      appClient.state.box.dataBoxes.value(metadata.start + BigInt(i))
    )
  );
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
      method: abiMethod,
      args: [pubKey, boxIndex, BYTES_PER_CALL * (i + bytesOffset), chunk],
      boxReferences: boxes,
      sender,
      appId: appID,
    });
  });
  await new Promise((r) => setTimeout(r, 500));
  return (await composer.send({ maxRoundsToWaitForConfirmation: 3 })).txIds;
}

/**
 *
 * @param atc
 * @param algodClient
 * @param retryCount
 */
async function tryExecute(composer: TransactionComposer, retryCount = 1): Promise<void> {
  try {
    await composer.send({ maxRoundsToWaitForConfirmation: 3 });
  } catch (e) {
    if (retryCount >= 3) throw e;
    await new Promise(r => setTimeout(r, 500 * retryCount));
    await tryExecute(composer, retryCount + 1);
  }
}

/**
 *
 * @param dataSize
 * @param boxCount
 * @param endBoxSize
 * @returns
 */
const calculateUploadCost = (boxCount: number, endBoxSize: number): number => {
  return (
    boxCount * COST_PER_BOX +
    (boxCount - 1) * MAX_BOX_SIZE * COST_PER_BYTE +
    boxCount * 8 * COST_PER_BYTE +
    endBoxSize * COST_PER_BYTE +
    COST_PER_BOX +
    (8 + 8 + 1 + 8 + 32 + 8) * COST_PER_BYTE
  );
};

/**
 * 
  * Splits data into boxes of MAX_BOX_SIZE bytes
 */
const splitDataIntoBoxes = (data: Buffer): Uint8Array[] => {
  const numBoxes = Math.floor(data.byteLength / MAX_BOX_SIZE);
  const boxes: Uint8Array[] = [];
  for (let i = 0; i < numBoxes; i++) {
    boxes.push(new Uint8Array(data.subarray(i * MAX_BOX_SIZE, (i + 1) * MAX_BOX_SIZE)));
  }
  // Push last box (may be smaller)
  boxes.push(new Uint8Array(data.subarray(numBoxes * MAX_BOX_SIZE, data.byteLength)));
  return boxes;
};

/**
 * Splits a box into chunks of BYTES_PER_CALL bytes
 */
const splitBoxIntoChunks = (box: Uint8Array): Uint8Array[] => {
  const chunks: Uint8Array[] = [];
  for (let i = 0; i < box.byteLength; i += BYTES_PER_CALL) {
    chunks.push(box.subarray(i, i + BYTES_PER_CALL));
  }
  return chunks;
};


/**
 * Uploads a DID document to the Algorand blockchain.
 * Steps:
 * 1. Calculate costs and prepare payment.
 * 2. Start upload and get metadata.
 * 3. Split data into boxes and chunks, upload each.
 * 4. Validate upload and finish.
 */

export async function uploadDIDDocument(
  appClient: DidAlgoStorageClient,
  data: Buffer,
  appID: bigint,
  pubKey: Uint8Array,
  sender: Address,
  algorand: AlgorandClient,
): Promise<Metadata> {
  // --- 1. Calculate costs and prepare payment ---
  const boxCount = Math.ceil(data.byteLength / MAX_BOX_SIZE);
  const endBoxSize = data.byteLength % MAX_BOX_SIZE;
  const totalCost = calculateUploadCost(boxCount, endBoxSize);
  const mbrPayment = appClient.algorand.createTransaction.payment({
    sender,
    receiver: appClient.appAddress,
    amount: microAlgos(totalCost),
  });

  // --- 2. Start upload and get metadata ---
  await appClient.send.startUpload({
    args: [algosdk.encodeAddress(pubKey), boxCount, endBoxSize, mbrPayment],
    boxReferences: [pubKey],
  });

  // Metadata for the upload so we know which boxes to write to
  const boxValue: any = await appClient.state.box.metadata.value(algosdk.encodeAddress(pubKey));
  const metadata: Metadata = {
    // index of first box
    start: boxValue.start, 
    // index of last box
    end: boxValue.end, 
    // upload status, it can be 0 (uploading), 1 (ready), 2 (deleting)
    status: boxValue.status,
    // we need to know how much data is in the last box, since we might not fill it completely
    // This is useful to then retrieve and validate the data
    endSize: boxValue.endSize,
  };

  // --- 3. Split data into boxes and chunks, upload each ---
  const boxData = splitDataIntoBoxes(data);

  const toRun: Promise<void>[] = boxData.map(async (box, boxIndexOffset) => {
    const boxIndex = metadata.start + BigInt(boxIndexOffset);

    // Even though a box can hold up to 32kb, we need to split the data for a box into smaller chunks
    // because each app call can only write up to ~2kb of data.
    const chunks = splitBoxIntoChunks(box);

    // Identify box references for this upload
    // 7 boxes for data + 1 box for metadata
    // note: 0n means current app ID
    const boxes: BoxReference[] = new Array(7)
      .fill({ appId: 0n, name: algosdk.encodeUint64(boxIndex) })
      .concat({ appId: 0n, name: pubKey });

    // upload chunks in groups of 8 (max number of app calls in a group) 
    await sendTxGroup(
      algorand,
      appClient.appClient.getABIMethod("upload")!,
      0,
      pubKey,
      boxes,
      boxIndex,
      sender,
      appID,
      chunks.slice(0, 8),
    );
    
    // The first 8 chunks have been sent, although there might be remaining chunks, specially at the end of the upload
    // This is the scenario where we are partially filling the last box
    if (chunks.length > 8) {
      await sendTxGroup(
        algorand,
        appClient.appClient.getABIMethod("upload")!,
        8,
        pubKey,
        boxes,
        boxIndex,
        sender,
        appID,
        chunks.slice(8),
      );
    }
  });

  await Promise.all(toRun);

  // --- 4. Validate upload and finish ---
  if (Buffer.concat(boxData).toString("hex") !== data.toString("hex")) {
    throw new Error("Data validation failed!");
  }

  // --- 5. Finish upload ---
  // update metadata to mark upload as complete
  await appClient.send.finishUpload({
    args: [algosdk.encodeAddress(pubKey)],
    boxReferences: [pubKey],
  });

  return metadata;
}

/**
 * Deletes a DID document from the Algorand blockchain.
 * Steps:
 * 1. Initialize AppClient and fetch metadata.
 * 2. Start the delete process.
 * 3. For each box, call deleteData and dummy methods to clear data.
 */

export async function deleteDIDDocument(
  appID: bigint,
  pubKey: Uint8Array,
  sender: Address,
  algorand: AlgorandClient,
): Promise<void> {
  // --- 1. Initialize AppClient and fetch metadata ---
  const appClient = new AppClient({
    appId: appID,
    defaultSender: algorand.account.random(),
    appSpec,
    algorand,
  });

  // Fetch box metadata (start, end, status, endSize)
  const boxValue = (
    await appClient.getBoxValueFromABIType(
      pubKey,
      algosdk.ABIType.from("(uint64,uint64,uint8,uint64,uint64)")
    )
  ).valueOf() as bigint[];
  const metadata: Metadata = {
    start: boxValue[0],
    end: boxValue[1],
    status: boxValue[2],
    endSize: boxValue[3],
  };

  // --- 2. Start the delete process ---
  await appClient.send.call({
    method: "startDelete",
    args: [pubKey],
    boxReferences: [pubKey],
    sender,
  });

  // --- 3. For each box, call deleteData and dummy methods to clear data ---
  for (let boxIndex = metadata.start; boxIndex <= metadata.end; boxIndex += 1n) {
    const composer = algorand.newGroup();
    const boxIndexRef = { appId: appID, name: algosdk.encodeUint64(boxIndex) };

    // Call deleteData ABI method for this box
    composer.addAppCallMethodCall({
      appId: appID,
      method: appClient.getABIMethod("deleteData")!,
      args: [pubKey, boxIndex],
      boxReferences: [
        { appId: appID, name: pubKey },
        ...Array(7).fill(boxIndexRef),
      ],
      extraFee: microAlgos(1000),
      sender,
    });

    // Call dummy ABI method 4 times (possibly to force box deletion)
    for (let i = 0; i < 4; i++) {
      composer.addAppCallMethodCall({
        appId: appID,
        method: appClient.getABIMethod("dummy")!,
        args: [],
        boxReferences: Array(8).fill(boxIndexRef),
        sender,
        note: new Uint8Array(Buffer.from(`dummy ${i}`)),
      });
    }

    // Execute the composed transaction group with retries
    await tryExecute(composer);
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
