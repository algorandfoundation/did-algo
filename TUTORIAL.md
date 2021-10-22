# Getting Started

This project provides the CLI-based application `algoid` which can be used to:

- Create and manage as many "Decentralized Identifiers (DIDs)" as you wish.
- Create and manage as many Algorand wallets (standalone) as you need.
- Connect any of your wallets to any of your DIDs.
- Publish the information about any of your DIDs for global access.

## 1. Installation

To get started you first need to install the latest version of `algoid` application,
there are 3 main installation mechanisms.

### 1.1 Using a release package (recommended)

You can download and use a pre-compiled release package from the official project repository.
The latest release can be [accessed here](https://github.com/algorandfoundation/did-algo/releases/latest).
There are a number of installers availabe, choose the proper one for your local system.
Officially supported OS / Architectures include:

- macOS with Intel chips (`darwin_amd64`)
- macOS with M1 chips (`darwin_arm64`)
- Windows 32bits (`windows_386`)
- Windows 64bits (`windows_amd64`)
- Debian-based Linux distributions on 32bits (`linux_386.deb`)
- Debian-based Linux distributions on 64bits (`linux_amd64.deb`)
- RedHat-based Linux distributions on 32bits (`linux_386.rpm`)
- RedHat-based Linux distributions on 64bits (`linux_amd64.rpm`)
- Alpine-based Linux distributions on 32bits (`linux_386.apk`)
- Alpine-based Linux distributions on 64bits (`linux_amd64.apk`)

For example, to install the application version `v0.1.0` on a 64bit machine running Ubuntu
Linux you can download the `did-algo_0.1.0_linux_amd64.deb` package and perform a installation
with your package manager tools.

```shell
dpkg -i did-algo_0.1.0_linux_amd64.deb
```

Once the process is complete you should be able to use the `algoid` tool; you can
verify the installation running `version` command.

```shell
algoid version
```

### 1.2 Using the installer script

- The script requires root or sudo privileges to move `algoid` binary to `/usr/local/bin`.
- The script attempts to detect your operating system (macOS or Linux) and architecture
  (arm64, x86, amd64) to download the appropriate binary from the releases page.
- By default, the script installs the latest version of `algoid`.
- The script requires packages `curl` and `sudo` to be already installed.

```shell
curl -sSfL https://raw.githubusercontent.com/algorandfoundation/did-algo/main/scripts/install.sh | sh
```

### 1.3 Building from source

To get the source and build the tools locally you need to:

1. If not already available, install a regular GoLang work environment following
   the [official instructions](https://golang.org/doc/install).

2. Using git, get the source code of the official `algoid` repository.

```shell
git clone https://github.com/algorandfoundation/did-algo.git
```

3. Once inside the working directory containing the source code, checkout a specific
   tagged revision to build and install. For example, to use `v0.1.0`.

```shell
  git checkout v0.1.0
```

4. Build the project artifacts.

```shell
make build
```

5. Install the artifacts to your local `$GOPATH/bin` directory. This location should
   be included as part of your global `$PATH`.

```shell
make install
```

6. Once the process is complete you should be able to use the `algoid` tool; you can
   verify the installation running `version` command.

```shell
algoid version
```

## 2. Basic Usage

Once `algoid` is installed you are ready to build your Digital Identity connected
to the Algorand Blockchain. The different functions available are executed as
individual commands and subcommands like `algoid <command> <subcommand>`; you can
get detailed information of every command or subcommand using the `-h` or `--help`
flag: `algoid <command> <subcommand> --help`.

To list all the main commands available simply run `algoid` without any additional
parameters.

```txt
Algorand DID

Reference client implementation for the "algo" DID method. The platform allows
entities to fully manage Decentralized Identifiers as described by version v1.0
of the specification.

For more information:
https://github.com/algorandfoundation/did-algo

Usage:
  algoid [command]

Available Commands:
  agent       Start a network agent supporting the DID method requirements
  completion  Generate autocompletion for commonly used shells
  delete      Permanently delete a local identifier
  edit        Edit local DIDs
  help        Help about any command
  info        Display the current information available on an existing DID
  list        List registered DIDs
  proof       Produce a linked digital proof document
  register    Creates a new DID locally
  retrieve    Retrieve the DID document of an existing identifier
  sync        Publish a DID instance to the processing network
  verify      Check the validity of a ProofLD document
  version     Display version information
  wallet      Manage your ALGO wallet(s)

Flags:
      --config string   config file ($HOME/.algoid/config.yaml)
  -h, --help            help for algoid
      --home string     home directory ($HOME/.algoid)

Use "algoid [command] --help" for more information about a command.
```

Let's explore some of the basic functions available.

## 3. DID Managent

Decentralized Identifiers are completely controlled by you, without coordination
with any central authority and can be securely used between peers on a trustless
environment via cryptographic verification. To this end, each DID instance can
include several "verification methods" and describe how each will be used in the
conext of different "verification relationships".

> For more information you can refer to the [latest specification](https://w3c.github.io/did-core/#verification-methods).

### 3.1 Create a new DID

To create a new DID, with a new passphrase-protected cryptographic key enabled
for authentication simply run:

```shell
algoid create my-first-did --passphrase
```

You'll be asked to enter and confirm you passphrase and finally get an output
similar to:

```shell
2021-10-19T22:19:33-05:00 INF generating new identifier method=algo
2021-10-19T22:19:33-05:00 DBG adding master key
2021-10-19T22:19:33-05:00 DBG setting master key as authentication mechanism
2021-10-19T22:19:33-05:00 INF adding entry to local store
```

### 3.2 List existing DIDs

As mentioned earlier, you can create as many DIDs as you wish. You can then
see a list of all your identifiers.

```shell
algoid list
```

The list displays every DID instance along it's local alias for simpler usage.

```txt
Name            DID
my-first-did    did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef
```

### 3.3 Inspect your local DID document

Every DID is resolvable to a DID document. A DID document provides all the
information required to validate control and ownership over a given identifier.

You can inspect the all the information associated with you identifier using
the `info` command.

```shell
algoid info my-first-did
```

The command will print on the screen the full contents of the associated DID
document.

```json
{
  "@context": [
    "https://www.w3.org/ns/did/v1",
    "https://w3id.org/security/v1"
  ],
  "id": "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef",
  "created": "2021-10-20T03:19:33Z",
  "updated": "2021-10-20T03:19:33Z",
  "verificationMethod": [
    {
      "id": "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef#master",
      "type": "Ed25519VerificationKey2018",
      "controller": "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef",
      "publicKeyMultibase": "zCh9PDTZzeWxk2WdH4M1e8k2951D5D11jz7Uti9HRBGiK"
    }
  ],
  "authentication": [
    "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef#master"
  ]
}
```

### 3.4 Edit a DID

The details associated with any of your identifiers can be modified using the
`edit` command. This includes adding or removing different verification methods
and/or service endpoints to disclose interaction capabilities.

```txt
Edit local DIDs

Usage:
  algoid edit [command]

Available Commands:
  key         Manage cryptographic keys associated with the DID
  service     Manage services enabled for the identifier
```

For example, to add a new cryptographic key to you identifier you can use the
`edit key add` command.

```txt
Add a new cryptographic key for the DID

Usage:
  algoid edit key add [flags]

Examples:
algoid edit key add [DID reference name] --name my-new-key --type ed --authentication

Flags:
  -a, --authentication   enable this key for authentication purposes
  -h, --help             help for add
  -n, --name string      name to be assigned to the newly added key (default "key-#")
  -t, --type string      type of cryptographic key: RSA (rsa), Ed25519 (ed) or secp256k1 (koblitz) (default "ed")
```

## 4 Wallet Management

The `algoid` application also provides all the tools you need to create and manage
any number of standalone Algorand wallets. You can explore all the functions available
by inspecting the `wallet` command.

```txt
Manage your ALGO wallet(s)

Usage:
  algoid wallet [command]

Available Commands:
  connect     Connect your ALGO wallet to a DID
  create      Create a new (standalone) ALGO wallet
  delete      Permanently delete an ALGO wallet
  disconnect  Remove a linked ALGO address from your DID
  export      Export wallet's master derivation key
  info        Get account information
  list        List your existing ALGO wallet(s)
  pay         Create and submit a new transaction
  rename      Rename an existing ALGO wallet
  restore     Restore a wallet using an existing mnemonic file
  watch       Monitor your wallet's activity
```

### 4.1 Create Wallet

Wallets are created and manage locally, the sensitive cryptographic materials
required to operate your locally wallet are encrypted prior to be written to the
local filesystem.

Let's create a new wallet, we'll use the `wallet create` command.

```shell
algoid wallet create sample-account
```

You'll be asked to enter and confirm a passphrase, this will be used as the
encryption key required for secure storage. Finally, you'll get an output similar
to this.

```shell
2021-10-19T22:52:11-05:00 INF new wallet created address=GZAURFQR7VRUL2TRVMUJ3G3IGFYSFXBWHBFFQBEITXWDLFVVHTICFNY2LE name=sample-account
```

### 4.2 List existing Wallets

As mentioned earlier, you can create as many wallets as you wish. You can then
see a list of all your local wallets.

```shell
algoid wallet list
```

The list displays every wallet using its local alias for simpler usage.

```txt
2021-10-19T22:55:58-05:00 INF wallet found: sample-account
```

### 4.3 Get wallet details

To get additional details such as your account balance, status, rewards, etc; simply
use the `wallet info` command.

```shell
algoid wallet info sample-account
```

The client application will reach out to the network and the account information will
be printed.

```txt
2021-10-19T22:57:58-05:00 INF establishing connection to network agent: did-agent.aidtech.network:443
address: GZAURFQR7VRUL2TRVMUJ3G3IGFYSFXBWHBFFQBEITXWDLFVVHTICFNY2LE
status: Offline
current balance: 0
pending rewards: 0
total rewards: 0
```

### 4.4 Connect a wallet to a DID

Now that you have a new DID instance and a new Algorand account, you can use
the `wallet connect` command to link both. This will enable discoverability and
interoperability for different services that require to link your ID to a
secure and performant digital payments channel.

```txt
Connect your ALGO wallet to a DID

Connecting your ALGO wallet to a DID will allow other users to
discover your ALGO address when resolving your identifier.

Effectively connecting your ID to a highly secure and efficient
payments channel. Additionally, your counterparties might also
discover or request your credentials when/if required to perform
certain transactions.

Usage:
  algoid wallet connect [flags]

Aliases:
  connect, link, ln

Examples:
algoid wallet connect [wallet-name] [did-name]
```

For example, to link the DID and created previously created.

```shell
algoid wallet connect sample-account my-first-did
```

After providing the wallet decryption key you'll get a confirmation message.

```txt
2021-10-19T23:07:55-05:00 INF updating local DID record
```

To confirm the connection you can inspect the details for your DID.

```shell
algoid info my-first-did
```

The updated DID document now include the `algo-connect` extension with
details about the connected Algorand account.

```json
{
  "@context": [
    "https://www.w3.org/ns/did/v1",
    "https://w3id.org/security/v1"
  ],
  "id": "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef",
  "created": "2021-10-20T03:19:33Z",
  "updated": "2021-10-20T04:07:55Z",
  "verificationMethod": [
    {
      "id": "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef#master",
      "type": "Ed25519VerificationKey2018",
      "controller": "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef",
      "publicKeyMultibase": "zCh9PDTZzeWxk2WdH4M1e8k2951D5D11jz7Uti9HRBGiK"
    }
  ],
  "authentication": [
    "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef#master"
  ],
  "service": [
    {
      "id": "did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef#algo-connect",
      "type": "did.algorand.foundation.ExternalService",
      "serviceEndpoint": "https://did.algorand.foundation",
      "extensions": [
        {
          "id": "algo-address",
          "version": "0.1.0",
          "data": [
            {
              "address": "GZAURFQR7VRUL2TRVMUJ3G3IGFYSFXBWHBFFQBEITXWDLFVVHTICFNY2LE",
              "asset": "ALGO",
              "network": "testnet"
            }
          ]
        }
      ]
    }
  ]
}
```

## 5. Publish your DID globally

Up to this point all the details about our DID are just available on our machine;
to be really useful we need to be able to publish this information on a decentralized
way so that others can retrieve it when required. To this end, the `algo` DID method
utilizes a robust integration with the [IPFS](https://ipfs.io/) decentralized storage
protocol.

To publish your local DID use the `publish` command.

```txt
Publish a DID instance to the processing network

Usage:
  algoid sync [flags]

Aliases:
  sync, publish, update, upload, push

Examples:
algoid sync [DID reference name]

Flags:
  -h, --help         help for sync
  -k, --key string   cryptographic key to use for the sync operation (default "master")
  -p, --pow int      set the required request ticket difficulty level (default 24)
```

For example, by running `algoid publish my-first-did` you'll get an output similar to
the following.

```txt
2021-10-19T23:18:56-05:00 DBG key selected for the operation: did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef#master
2021-10-19T23:18:56-05:00 INF publishing: my-first-did
2021-10-19T23:18:56-05:00 INF generating request ticket pow=16
2021-10-19T23:18:56-05:00 DBG ticket obtained: 00359dea149f5de57be799a5f953bb46e8640011c8ec2978f6855f17d1efcbfa
2021-10-19T23:18:56-05:00 DBG time: 1.081207ms (rounds completed 15)
2021-10-19T23:18:56-05:00 INF establishing connection to network agent: did-agent.aidtech.network:443
2021-10-19T23:18:56-05:00 INF submitting request to the network
2021-10-19T23:18:58-05:00 DBG request status: true
2021-10-19T23:18:58-05:00 INF identifier: /ipfs/bafkreicws5lv7mu335gc7rkr4wkvmhkqftlvimm5fd3piupwnlinnmlzp4
```

Notice the IPFS CID generated: `/ipfs/bafkreicws5lv7mu335gc7rkr4wkvmhkqftlvimm5fd3piupwnlinnmlzp4`;
this identifier can be used to retrieve the now globally published DID document from any publicly
available IPFS Gateway, like [dweb.link](https://dweb.link/ipfs/bafkreicws5lv7mu335gc7rkr4wkvmhkqftlvimm5fd3piupwnlinnmlzp4).

### 5.1 Resolve a DID

Finally, other should be able to easily retrieve the DID document associated with
any particular DID instance. The `algoid` application provides the `resolve` command
for this specific purpose.

```txt
Retrieve the DID document of an existing identifier

Usage:
  algoid retrieve [flags]

Aliases:
  retrieve, get, resolve

Examples:
algoid retrieve [existing DID]
```

For example, to resolve the DID created as part of this tutorial.

```shell
algoid resolve did:algo:f30304e2-e702-4f47-b242-2e9e064f6eef
```
