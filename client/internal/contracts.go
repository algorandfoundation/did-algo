package internal

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"

	"github.com/algorand/go-algorand-sdk/v2/abi"
)

// References:
// https://developer.algorand.org/docs/get-details/encoding/

// StorageContracts contains the pre-compiled smart contracts to
// support AlgoDID's on-chain storage.
var StorageContracts fs.FS

//go:embed contracts
var dist embed.FS

var (
	approvalTeal []byte
	clearTeal    []byte
)

func init() {
	StorageContracts, _ = fs.Sub(dist, "contracts")

	// load approval program
	approvalFile, err := StorageContracts.Open("DIDAlgoStorage.approval.teal")
	if err != nil {
		panic(err)
	}
	approvalTeal, err = io.ReadAll(approvalFile)
	if err != nil {
		panic(err)
	}
	_ = approvalFile.Close()

	// load clear program
	clearFile, err := StorageContracts.Open("DIDAlgoStorage.clear.teal")
	if err != nil {
		panic(err)
	}
	clearTeal, err = io.ReadAll(clearFile)
	if err != nil {
		panic(err)
	}
	_ = clearFile.Close()
}

// LoadContract loads the AlgoDID smart contract ABI from JSON file.
func LoadContract() *abi.Contract {
	abiFile, _ := StorageContracts.Open("DIDAlgoStorage.arc4.json")
	abiContents, _ := io.ReadAll(abiFile)
	contract := &abi.Contract{}
	_ = json.Unmarshal(abiContents, contract)
	_ = abiFile.Close()
	return contract
}
