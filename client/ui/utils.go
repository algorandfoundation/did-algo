package ui

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/did"
	"go.bryk.io/x/crypto/ed25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
)

// Securely expand the provided secret material.
func expand(secret []byte, size int, info []byte) ([]byte, error) {
	salt := make([]byte, sha256.Size)
	buf := make([]byte, size)
	h := hkdf.New(sha3.New256, secret, salt[:], info)
	if _, err := io.ReadFull(h, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// Restore key pair from the provided material.
func keyFromMaterial(material []byte) (*ed25519.KeyPair, error) {
	m, err := expand(material, 32, nil)
	if err != nil {
		return nil, err
	}
	seed := [32]byte{}
	copy(seed[:], m)
	return ed25519.FromSeed(seed[:])
}

// Return the list of ALGO addresses linked to the provided identifier.
func getAlgoAddress(id *did.Identifier) []addressEntry {
	var result = []addressEntry{}
	svc := id.Service("algo-connect")
	if svc == nil {
		return result
	}
	var addresses []algoDestination
	ext := did.Extension{
		ID:      "algo-address",
		Version: "0.1.0",
	}
	if err := svc.GetExtension(ext.ID, ext.Version, &addresses); err != nil {
		return result
	}
	for _, entry := range addresses {
		result = append(result, addressEntry{
			Address: entry.Address,
			Network: entry.Network,
			Enabled: true,
		})
	}
	return result
}

// Get the key used for the sync operation.
func getSyncKey(id *did.Identifier) (*did.VerificationKey, error) {
	// Get selected key for the sync operation
	key := id.VerificationMethod("master")
	if key == nil {
		return nil, errors.New("invalid key selected")
	}

	// Verify the key is enabled for authentication
	isAuth := false
	for _, k := range id.GetVerificationRelationship(did.AuthenticationVM) {
		if k == key.ID {
			isAuth = true
			break
		}
	}
	if !isAuth {
		return nil, errors.New("the key selected is not enabled for authentication purposes")
	}
	return key, nil
}

// Generate a new sync request.
func getRequestTicket(id *did.Identifier, key *did.VerificationKey) (*protoV1.Ticket, error) {
	diff := uint(viper.GetInt("client.pow"))

	// Create new ticket
	ticket, err := protoV1.NewTicket(id, key.ID)
	if err != nil {
		return nil, err
	}

	// Solve PoW challenge and sign ticket
	challenge := ticket.Solve(context.Background(), diff)
	ch, _ := hex.DecodeString(challenge)
	if ticket.Signature, err = key.Sign(ch); err != nil {
		return nil, fmt.Errorf("failed to generate request ticket: %w", err)
	}

	// Verify on client's side
	if err = ticket.Verify(diff); err != nil {
		return nil, fmt.Errorf("failed to verify ticket: %w", err)
	}
	return ticket, nil
}

// Return an empty service entry for `algo-connect`.
func newServiceEntry() *did.ServiceEndpoint {
	return &did.ServiceEndpoint{
		ID:       "algo-connect",
		Type:     "AlgorandExternalService",
		Endpoint: "https://did.algorand.foundation",
		Extensions: []did.Extension{
			{
				ID:      "algo-address",
				Version: "0.1.0",
				Data:    []algoDestination{},
			},
		},
	}
}
