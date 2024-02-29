// nolint
package internal

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/algorand/go-algorand-sdk/v2/abi"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"
)

const (
	cost_per_byte  = 400
	cost_per_box   = 2500
	max_box_size   = 32768
	bytes_per_call = 2048 - 4 - 34 - 8 - 8
)

// CreateApp is used to deploy the AlgoDID storage smart contract to the
// Algorand network.
func createApp(
	algodClient *algod.Client,
	contract *abi.Contract,
	sender types.Address,
	signer transaction.TransactionSigner,
) (uint64, error) {
	atc := transaction.AtomicTransactionComposer{}

	// Grab the method from out contract object
	method, err := contract.GetMethodByName("createApplication")
	if err != nil {
		return 0, fmt.Errorf("failed to get add method: %w", err)
	}

	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get suggested params: %w", err)
	}

	compiledApproval, err := algodClient.TealCompile(approvalTeal).Do(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to compile approval: %w", err)
	}

	compiledClear, err := algodClient.TealCompile(clearTeal).Do(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to compile clear: %w", err)
	}

	approvalProgram, err := base64.StdEncoding.DecodeString(compiledApproval.Result)
	if err != nil {
		return 0, fmt.Errorf("failed to decode approval program: %w", err)
	}

	clearProgram, err := base64.StdEncoding.DecodeString(compiledClear.Result)
	if err != nil {
		return 0, fmt.Errorf("failed to decode clear program: %w", err)
	}

	mcp := transaction.AddMethodCallParams{
		AppID:           0,
		Sender:          sender,
		SuggestedParams: sp,
		OnComplete:      types.NoOpOC,
		Signer:          signer,
		Method:          method,
		MethodArgs:      []interface{}{},
		ApprovalProgram: approvalProgram,
		ClearProgram:    clearProgram,
		GlobalSchema:    types.StateSchema{NumUint: 1},
	}
	if err := atc.AddMethodCall(mcp); err != nil {
		return 0, fmt.Errorf("failed to add method call: %w", err)
	}

	result, err := atc.Execute(algodClient, context.Background(), 3)
	if err != nil {
		return 0, fmt.Errorf("failed to execute atomic transaction: %w", err)
	}

	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, result.TxIDs[0], 4, context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to wait for confirmation: %w", err)
	}

	appID := confirmedTxn.ApplicationIndex
	fundAtc := transaction.AtomicTransactionComposer{}
	mbrPayment, err := transaction.MakePaymentTxn(
		sender.String(),
		crypto.GetApplicationAddress(appID).String(),
		uint64(100_000),
		nil,
		"",
		sp,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to make payment txn: %w", err)
	}

	var mbrPaymentWithSigner transaction.TransactionWithSigner
	mbrPaymentWithSigner.Txn = mbrPayment
	mbrPaymentWithSigner.Signer = signer
	if err = fundAtc.AddTransaction(mbrPaymentWithSigner); err != nil {
		return 0, fmt.Errorf("failed to add transaction: %w", err)
	}
	if _, err = fundAtc.Execute(algodClient, context.Background(), 3); err != nil {
		return 0, fmt.Errorf("failed to execute atomic transaction: %w", err)
	}
	return confirmedTxn.ApplicationIndex, nil
}

// PublishDID is used to upload a new DID document to the AlgoDID
// storage smart contract.
func publishDID(
	algodClient *algod.Client,
	appID uint64,
	contract *abi.Contract,
	sender types.Address,
	signer transaction.TransactionSigner,
	data []byte,
	pubKey []byte,
) error {
	ceilBoxes := int(math.Ceil(float64(len(data)) / float64(max_box_size)))
	endBoxSize := len(data) % max_box_size

	totalCost := ceilBoxes*cost_per_box + (ceilBoxes-1)*max_box_size*cost_per_byte + ceilBoxes*8*cost_per_byte + endBoxSize*cost_per_byte + cost_per_box + (8+8+1+8+32+8)*cost_per_byte

	atc := transaction.AtomicTransactionComposer{}

	// Grab the method from out contract object
	method, err := contract.GetMethodByName("startUpload")
	if err != nil {
		return fmt.Errorf("failed to get add method: %w", err)
	}

	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get suggested params: %w", err)
	}

	mbrPayment, err := transaction.MakePaymentTxn(
		sender.String(),
		crypto.GetApplicationAddress(appID).String(),
		uint64(totalCost),
		nil,
		"",
		sp,
	)
	if err != nil {
		return fmt.Errorf("failed to make payment txn: %w", err)
	}
	var mbrPaymentWithSigner transaction.TransactionWithSigner
	mbrPaymentWithSigner.Txn = mbrPayment
	mbrPaymentWithSigner.Signer = signer

	byteType, err := abi.TypeOf("address")
	if err != nil {
		return fmt.Errorf("failed to get type of address: %w", err)
	}

	pubKeyAbiValue, err := byteType.Encode(pubKey)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	boxRefs := []types.AppBoxReference{{AppID: appID, Name: pubKey}}
	mcp := transaction.AddMethodCallParams{
		AppID:           appID,
		Sender:          sender,
		SuggestedParams: sp,
		OnComplete:      types.NoOpOC,
		Signer:          signer,
		Method:          method,
		BoxReferences:   boxRefs,
		MethodArgs:      []interface{}{pubKeyAbiValue, ceilBoxes, endBoxSize, mbrPaymentWithSigner},
	}
	if err := atc.AddMethodCall(mcp); err != nil {
		return fmt.Errorf("failed to add method call: %w", err)
	}

	_, err = atc.Execute(algodClient, context.Background(), 3)
	if err != nil {
		return fmt.Errorf("failed to execute atomic transaction: %w", err)
	}

	metadata, err := getMetadata(appID, pubKey, algodClient)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	numBoxes := int(math.Floor(float64(len(data)) / float64(max_box_size)))
	boxData := [][]byte{}
	for i := 0; i < numBoxes; i++ {
		upperBound := (i + 1) * max_box_size
		if len(data) < upperBound {
			upperBound = len(data)
		}
		box := data[i*max_box_size : upperBound]
		boxData = append(boxData, box)
	}

	// add data for the last box
	if len(data) > max_box_size {
		boxData = append(boxData, data[numBoxes*max_box_size:])
	}

	for boxIndexOffset, box := range boxData {
		boxIndex := metadata.Start + uint64(boxIndexOffset)
		encodedBoxIndex := make([]byte, 8)
		binary.BigEndian.PutUint64(encodedBoxIndex, boxIndex)
		numChunks := int(math.Ceil(float64(len(box)) / float64(bytes_per_call)))
		chunks := [][]byte{}
		for i := 0; i < numChunks; i++ {
			upperBound := (i + 1) * bytes_per_call
			if len(box) < upperBound {
				upperBound = len(box)
			}
			chunks = append(chunks, box[i*bytes_per_call:upperBound])
		}

		boxRef := types.AppBoxReference{AppID: appID, Name: encodedBoxIndex}
		boxes := []types.AppBoxReference{}
		for i := 0; i < 7; i++ {
			boxes = append(boxes, boxRef)
		}

		boxes = append(boxes, types.AppBoxReference{AppID: appID, Name: pubKey})

		uploadMethod, err := contract.GetMethodByName("upload")
		if err != nil {
			return fmt.Errorf("failed to get add method: %w", err)
		}

		_, err = sendTxGroup(algodClient, uploadMethod, 0, pubKey, boxes, boxIndex, sp, sender, signer, appID, chunks[:8])
		if err != nil {
			return fmt.Errorf("failed to send tx group: %w", err)
		}

		if numChunks > 8 {
			_, err = sendTxGroup(algodClient, uploadMethod, 8, pubKey, boxes, boxIndex, sp, sender, signer, appID, chunks[8:])
			if err != nil {
				return fmt.Errorf("failed to send tx group: %w", err)
			}
		}
	}

	finishUploadMethod, err := contract.GetMethodByName("finishUpload")
	if err != nil {
		return fmt.Errorf("failed to get add method: %w", err)
	}

	finishUploadMcp := transaction.AddMethodCallParams{
		AppID:           appID,
		Sender:          sender,
		SuggestedParams: sp,
		OnComplete:      types.NoOpOC,
		Signer:          signer,
		Method:          finishUploadMethod,
		BoxReferences:   []types.AppBoxReference{{AppID: appID, Name: pubKey}},
		MethodArgs:      []interface{}{pubKey},
	}

	finishAtc := transaction.AtomicTransactionComposer{}
	if err := finishAtc.AddMethodCall(finishUploadMcp); err != nil {
		return fmt.Errorf("failed to add method call: %w", err)
	}

	_, err = finishAtc.Execute(algodClient, context.Background(), 3)
	return err
}

// ResolveDID is used to read the DID document from the AlgoDID storage smart
// contract.
func resolveDID(appID uint64, pubKey []byte, algodClient *algod.Client) ([]byte, error) {
	metadata, err := getMetadata(appID, pubKey, algodClient)
	if err != nil {
		return nil, err
	}
	if metadata.Status == 0 {
		return nil, fmt.Errorf("DID document is being created")
	}

	if metadata.Status == 2 {
		return nil, fmt.Errorf("DID document is being deleted")
	}

	data := []byte{}
	for i := metadata.Start; i <= metadata.End; i++ {
		encodedBoxIndex := make([]byte, 8)
		binary.BigEndian.PutUint64(encodedBoxIndex, i)
		boxValue, err := algodClient.GetApplicationBoxByName(appID, encodedBoxIndex).Do(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to read box: %w", err)
		}
		data = append(data, boxValue.Value...)
	}

	return data, nil
}

// DeleteDID is used to delete the DID document from the AlgoDID
// storage smart contract.
func deleteDID(
	appID uint64,
	pubKey []byte,
	sender types.Address,
	algodClient *algod.Client,
	contract *abi.Contract,
	signer transaction.TransactionSigner,
) error {
	startAtc := transaction.AtomicTransactionComposer{}

	method, err := contract.GetMethodByName("startDelete")
	if err != nil {
		return fmt.Errorf("failed to get add method: %w", err)
	}
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get suggested params: %w", err)
	}
	byteType, err := abi.TypeOf("address")
	if err != nil {
		return fmt.Errorf("failed to get type of address: %w", err)
	}

	pubKeyAbiValue, err := byteType.Encode(pubKey)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	mcp := transaction.AddMethodCallParams{
		AppID:           appID,
		Sender:          sender,
		SuggestedParams: sp,
		OnComplete:      types.NoOpOC,
		Signer:          signer,
		Method:          method,
		BoxReferences:   []types.AppBoxReference{{AppID: appID, Name: pubKey}},
		MethodArgs:      []interface{}{pubKeyAbiValue},
	}
	if err := startAtc.AddMethodCall(mcp); err != nil {
		return fmt.Errorf("failed to add method call: %w", err)
	}
	if _, err = startAtc.Execute(algodClient, context.Background(), 3); err != nil {
		return fmt.Errorf("failed to execute atomic transaction: %w", err)
	}

	metadata, err := getMetadata(appID, pubKey, algodClient)
	if err != nil {
		return err
	}
	atcs := []struct {
		boxIndex uint64
		atc      transaction.AtomicTransactionComposer
	}{}

	for boxIndex := metadata.Start; boxIndex <= metadata.End; boxIndex++ {
		atc := transaction.AtomicTransactionComposer{}
		encodedBoxIndex := make([]byte, 8)
		binary.BigEndian.PutUint64(encodedBoxIndex, boxIndex)
		boxIndexRef := types.AppBoxReference{AppID: appID, Name: encodedBoxIndex}
		deleteDataMethod, err := contract.GetMethodByName("deleteData")
		if err != nil {
			return fmt.Errorf("failed to get method: %w", err)
		}

		sp.Fee = 2000
		sp.FlatFee = true
		if err := atc.AddMethodCall(transaction.AddMethodCallParams{
			AppID:           appID,
			Sender:          sender,
			SuggestedParams: sp,
			OnComplete:      types.NoOpOC,
			Signer:          signer,
			Method:          deleteDataMethod,
			BoxReferences:   []types.AppBoxReference{{AppID: appID, Name: pubKey}, boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef},
			MethodArgs:      []interface{}{pubKey, boxIndex},
		}); err != nil {
			return fmt.Errorf("failed to add method call: %w", err)
		}

		dummyMethod, err := contract.GetMethodByName("dummy")
		if err != nil {
			return fmt.Errorf("failed to get method: %w", err)
		}

		for i := 0; i < 4; i++ {
			if err := atc.AddMethodCall(transaction.AddMethodCallParams{
				AppID:           appID,
				Sender:          sender,
				SuggestedParams: sp,
				OnComplete:      types.NoOpOC,
				Signer:          signer,
				Method:          dummyMethod,
				BoxReferences:   []types.AppBoxReference{boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef, boxIndexRef},
				MethodArgs:      []interface{}{},
				Note:            []byte(fmt.Sprintf("dummy %d", i)),
			}); err != nil {
				return fmt.Errorf("failed to add method call: %w", err)
			}
		}

		atcs = append(atcs, struct {
			boxIndex uint64
			atc      transaction.AtomicTransactionComposer
		}{boxIndex, atc})
	}
	for _, atc := range atcs {
		if _, err = atc.atc.Execute(algodClient, context.Background(), 3); err != nil {
			return fmt.Errorf("failed to execute atomic transaction: %w", err)
		}
	}
	return nil
}

func getMetadata(appID uint64, pubKey []byte, algodClient *algod.Client) (metadata, error) {
	boxValue, err := algodClient.GetApplicationBoxByName(appID, pubKey).Do(context.Background())
	if err != nil {
		return metadata{}, fmt.Errorf("failed to get box: %w", err)
	}

	metadataType, err := abi.TypeOf("(uint64,uint64,uint8,uint64,uint64)")
	if err != nil {
		return metadata{}, fmt.Errorf("failed to get type of metadata: %w", err)
	}

	md, err := metadataType.Decode(boxValue.Value)
	if err != nil {
		return metadata{}, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return metadata{
		Start:   md.([]interface{})[0].(uint64),
		End:     md.([]interface{})[1].(uint64),
		Status:  md.([]interface{})[2].(uint8),
		EndSize: md.([]interface{})[3].(uint64),
	}, nil
}

func sendTxGroup(
	algodClient *algod.Client,
	abiMethod abi.Method,
	bytesOffset int,
	pubKey []byte,
	boxes []types.AppBoxReference,
	boxIndex uint64,
	suggestedParams types.SuggestedParams,
	sender types.Address,
	signer transaction.TransactionSigner,
	appID uint64,
	group [][]byte,
) ([]string, error) {
	atc := transaction.AtomicTransactionComposer{}

	for i, chunk := range group {
		atc.AddMethodCall(transaction.AddMethodCallParams{
			Method:          abiMethod,
			MethodArgs:      []interface{}{pubKey, boxIndex, bytes_per_call * (i + bytesOffset), chunk},
			BoxReferences:   boxes,
			SuggestedParams: suggestedParams,
			Sender:          sender,
			Signer:          signer,
			AppID:           appID,
		})
	}

	result, err := atc.Execute(algodClient, context.Background(), 3)
	if err != nil {
		return nil, fmt.Errorf("failed to execute atomic transaction: %w", err)
	}

	return result.TxIDs, nil
}

type metadata struct {
	Start   uint64
	End     uint64
	Status  uint8
	EndSize uint64
}
