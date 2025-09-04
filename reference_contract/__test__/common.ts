/* eslint-disable func-names */
import algosdk from 'algosdk';

export const indexerClient = new algosdk.Indexer('a'.repeat(64), 'http://localhost', 8980);
export const algodClient = new algosdk.Algodv2('a'.repeat(64), 'http://localhost', 4001);
export const kmdClient = new algosdk.Kmd('a'.repeat(64), 'http://localhost', 4002);
