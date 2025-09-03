import { AlgorandClient } from "@algorandfoundation/algokit-utils";
import express from "express";
import algosdk from "algosdk";
import { AppClient } from "@algorandfoundation/algokit-utils/types/app-client";
import appSpecJson from "../reference_contract/contracts/artifacts/DIDAlgoStorage.arc56.json";

const appSpec = JSON.stringify(appSpecJson);

export type Metadata = {
  start: bigint;
  end: bigint;
  status: bigint;
  endSize: bigint;
};

const app = express();
const port = 8080;

app.get("/1.0/identifiers/:identifier", async (req, res) => {
  const splitDid = req.params.identifier.split(":");

  const idxOffset = splitDid.length === 6 ? 0 : 1;

  const network = splitDid.length === 6 ? splitDid[2] : "mainnet";

  if (splitDid[0] !== "did") {
    res
      .status(400)
      .send(`invalid protocol, expected 'did', got ${splitDid[0]}`);
    return;
  }
  if (splitDid[1] !== "algo") {
    res
      .status(400)
      .send(`invalid DID method, expected 'algo', got ${splitDid[1]}`);
    return;
  }

  let algorand: AlgorandClient;

  switch (network) {
    case "mainnet":
      algorand = AlgorandClient.mainNet();
      break;
    case "testnet":
      algorand = AlgorandClient.testNet();
      break;
    case "custom":
      algorand = AlgorandClient.defaultLocalNet();
      break;
    default:
      throw Error(`Unsupported network: ${network}`);
  }

  const nameSpace = splitDid[3 - idxOffset];

  if (nameSpace !== "app") {
    res.status(400).send(`invalid namespace, expected 'app', got ${nameSpace}`);
    return;
  }

  const pubKeyHex = splitDid[5 - idxOffset];
  const pubKey = Buffer.from(pubKeyHex!, "hex");

  let appID: bigint;

  try {
    appID = BigInt(splitDid[4 - idxOffset]!);
    algosdk.encodeUint64(appID);
  } catch (e) {
    res
      .status(400)
      .send(`invalid app ID, expected uint64, got ${splitDid[4 - idxOffset]}`);
    return;
  }

  const appClient = new AppClient({
    appId: appID,
    defaultSender: algorand.account.random(),
    appSpec,
    algorand,
  });

  let metadata: Metadata;
  try {
    const metadataBoxValue = (
      await appClient.getBoxValueFromABIType(
        pubKey,
        algosdk.ABIType.from("(uint64,uint64,uint8,uint64,uint64)"),
      )
    ).valueOf() as bigint[];

    metadata = {
      start: metadataBoxValue[0]!,
      end: metadataBoxValue[1]!,
      status: metadataBoxValue[2]!,
      endSize: metadataBoxValue[3]!,
    };
  } catch (e) {
    res
      .status(404)
      .send(
        `Failed to get metadata from box. Ensure network (${network}), app ID (${appID}), and pubkey (${pubKeyHex}) are correct: ${e}`,
      );
    return;
  }

  if (metadata.status === BigInt(0)) {
    res.status(400).send("DID document is still being uploaded");
    return;
  }
  if (metadata.status === BigInt(2)) {
    res.status(400).send("DID document is being deleted");
    return;
  }

  const boxPromises = [];
  for (let i = metadata.start; i <= metadata.end; i += 1n) {
    boxPromises.push(appClient.getBoxValue(algosdk.encodeUint64(i)));
  }

  const boxValues = await Promise.all(boxPromises);

  const documentBytes = Buffer.concat(boxValues);

  const accept = req.get("Accept") ?? "application/did";

  console.debug(accept);

  const supportedContentTypes = ["did", "did+ld+json", "did+json", "json"].map(
    (t) => `application/${t}`,
  );

  switch (accept) {
    case "application/did":
    case "application/did+ld+json":
    case "application/did+json":
    case "application/json":
      try {
        const body = JSON.parse(documentBytes.toString());
        res.writeHead(200, { "Content-Type": "application/did" });
        res.json(body);
      } catch (e) {
        res.status(400).send(`Invalid JSON: ${e} `);
      }
      break;
    default:
      res
        .status(406)
        .send(
          `representation not supported: ${accept}. Supported representations: ${supportedContentTypes.join(", ")}`,
        );
  }
});

app.listen(port, () => {
  console.log(`Listening on port ${port}...`);
});
