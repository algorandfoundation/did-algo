#!/bin/bash

# Automatically exit on error
set -e

# Default values
INDEXER="no"
ARCHIVAL="no"
RELAY="no"

# can be set via flags or ENV variable
NETWORK=${NETWORK:-'testnet'}
TOKEN=${TOKEN:-''}

# Parse CLI flags
POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    -n|--network)
      NETWORK="$2"
      shift # past argument
      shift # past value
      ;;
    -t|--token)
      TOKEN="$2"
      shift # past argument
      shift # past value
      ;;
    -a|--archival)
      ARCHIVAL=yes
      shift # past argument
      ;;
    -i|--indexer)
      INDEXER=yes
      shift # past argument
      ;;
    -r|--relay)
      RELAY=yes
      shift # past argument
      ;;
    -h|--help)
      echo "Start an Algorand node instance"
      echo "  -n | --network    Network id (default testnet)"
      echo "  -t | --token      API access token"
      echo "  -r | --relay      Enable 'relay' mode"
      echo "  -a | --archival   Enable 'archival' mode"
      echo "  -i | --indexer    Enable 'indexer' mode"
      exit 0
      ;;
    *) # unknown option
      POSITIONAL+=("$1") # save it in an array for later
      shift # past argument
      ;;
  esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

# Validate network
if [ ! -f /var/lib/algorand/genesis/${NETWORK}/genesis.json ]; then
  echo "invalid network provided: ${NETWORK}"
  exit 1
fi

# Install genesis file if not present
if [ ! -f ${ALGORAND_DATA}/genesis.json ]; then
  cp /var/lib/algorand/genesis/${NETWORK}/genesis.json ${ALGORAND_DATA}/genesis.json
fi

# Install config files if not present
if [ ! -f ${ALGORAND_DATA}/system.json ]; then
  cp /var/lib/algorand/system.json ${ALGORAND_DATA}/system.json
fi
if [ ! -f ${ALGORAND_DATA}/config.json ]; then
  cp /var/lib/algorand/config.json ${ALGORAND_DATA}/config.json
fi

# API token
if [ -n ${TOKEN} ]; then
  echo ${TOKEN} > ${ALGORAND_DATA}/algod.token
  echo ${TOKEN} > ${ALGORAND_DATA}/algod.admin.token
fi

# Enable archival mode
if [ ${ARCHIVAL} == "yes" ]; then
  sed -i 's/"Archival": false/"Archival": true/g' ${ALGORAND_DATA}/config.json
fi

# Enable indexer mode
if [ ${INDEXER} == "yes" ]; then
  sed -i 's/"IsIndexerActive": false/"IsIndexerActive": true/g' ${ALGORAND_DATA}/config.json
fi

# Enable relay mode
if [ ${RELAY} == "yes" ]; then
  sed -i 's/"ForceRelayMessages": false/"ForceRelayMessages": true/g' ${ALGORAND_DATA}/config.json
fi

# Special betanet DNS settings
# https://developer.algorand.org/docs/run-a-node/operations/switch_networks/#dns-configuration-for-betanet
if [ ${NETWORK} == "betanet" ]; then
  sed -i 's/<network>.algorand.network/<network>.algodev.network/g' ${ALGORAND_DATA}/config.json
fi

# Start node
echo "Connecting to network: ${NETWORK}"
echo "Archival mode: ${ARCHIVAL}"
echo "Indexer mode: ${INDEXER}"

# Use 'exec' so that the 'algod' application becomes the container’s PID 1. This
# allows the application to receive any Unix signals sent to the container.
exec algod
