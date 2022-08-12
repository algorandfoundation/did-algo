package protov1

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.bryk.io/pkg/crypto/pow"
	"go.bryk.io/pkg/did"
	"golang.org/x/crypto/sha3"
)

const defaultTicketDifficultyLevel = 24

// NewTicket returns a properly initialized new ticket instance.
func NewTicket(id *did.Identifier, keyID string) (*Ticket, error) {
	// Get safe DID document
	contents, err := json.Marshal(id.Document(true))
	if err != nil {
		return nil, err
	}

	// Get proof
	proof, err := id.GetProof(keyID, "did.algorand.foundation")
	if err != nil {
		return nil, err
	}
	proofBytes, err := json.Marshal(proof)
	if err != nil {
		return nil, err
	}

	// Create new ticket
	t := &Ticket{
		Timestamp:  time.Now().UTC().Unix(),
		NonceValue: 0,
		KeyId:      keyID,
		Document:   contents,
		Proof:      proofBytes,
		Signature:  nil,
	}

	// Add metadata, if available
	if metadata := id.GetMetadata(); metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		t.DocumentMetadata = metadataBytes
	}

	return t, nil
}

// GetDID retrieve the DID instance from the ticket contents.
func (t *Ticket) GetDID() (*did.Identifier, error) {
	// Restore id instance from document
	doc := &did.Document{}
	if err := json.Unmarshal(t.Document, doc); err != nil {
		return nil, errors.New("invalid ticket contents")
	}
	id, err := did.FromDocument(doc)
	if err != nil {
		return nil, err
	}

	// Restore metadata, if available
	if t.DocumentMetadata != nil {
		metadata := &did.DocumentMetadata{}
		if err := json.Unmarshal(t.DocumentMetadata, metadata); err != nil {
			return nil, errors.New("invalid ticket contents")
		}
		if err := id.AddMetadata(metadata); err != nil {
			return nil, err
		}
	}
	return id, nil
}

// GetProofLD returns the decoded proof document contained in the ticket.
func (t *Ticket) GetProofLD() (*did.ProofLD, error) {
	proof := &did.ProofLD{}
	if err := json.Unmarshal(t.Proof, proof); err != nil {
		return nil, errors.New("invalid proof contents")
	}
	return proof, nil
}

// ResetNonce returns the internal nonce value back to 0.
func (t *Ticket) ResetNonce() {
	t.NonceValue = 0
}

// IncrementNonce will adjust the internal nonce value by 1.
func (t *Ticket) IncrementNonce() {
	t.NonceValue++
}

// Nonce returns the current value set on the nonce attribute.
func (t *Ticket) Nonce() int64 {
	return t.NonceValue
}

// MarshalBinary returns a deterministic binary encoding for the ticket
// instance using a byte concatenation of the form:
// 'timestamp | nonce | key_id | document | proof | document_metadata'
// where timestamp and nonce are individually encoded using little endian
// byte order.
func (t *Ticket) MarshalBinary() ([]byte, error) {
	var tc []byte
	nb := bytes.NewBuffer(nil)
	tb := bytes.NewBuffer(nil)
	kb := make([]byte, hex.EncodedLen(len([]byte(t.KeyId))))
	if err := binary.Write(nb, binary.LittleEndian, t.Nonce()); err != nil {
		return nil, fmt.Errorf("failed to encode nonce value: %w", err)
	}
	if err := binary.Write(tb, binary.LittleEndian, t.GetTimestamp()); err != nil {
		return nil, fmt.Errorf("failed to encode nonce value: %w", err)
	}
	hex.Encode(kb, []byte(t.KeyId))
	tc = append(tc, tb.Bytes()...)
	tc = append(tc, nb.Bytes()...)
	tc = append(tc, kb...)
	tc = append(tc, t.Document...)
	tc = append(tc, t.Proof...)
	tc = append(tc, t.DocumentMetadata...)
	return tc, nil

	// A simpler encoding mechanism using the standard protobuf encoder.
	// tc := proto.Clone(t).(*Ticket)
	// tc.Signature = nil
	// return proto.MarshalOptions{Deterministic: true}.Marshal(tc)
}

// Solve the ticket challenge using the proof-of-work mechanism.
func (t *Ticket) Solve(ctx context.Context, difficulty uint) string {
	if difficulty == 0 {
		difficulty = defaultTicketDifficultyLevel
	}
	return <-pow.Solve(ctx, t, sha3.New256(), difficulty)
}

// Verify perform all the required validations to ensure the request ticket is
// ready for further processing
//   - Challenge is valid
//   - Contents are a properly encoded DID instance
//   - Contents don’t include any private key, for security reasons no private keys should
//     ever be published on the network
//   - DID proof is valid
//   - Ticket signature is valid.
func (t *Ticket) Verify(difficulty uint) error {
	// Challenge is valid
	if difficulty == 0 {
		difficulty = defaultTicketDifficultyLevel
	}
	if !pow.Verify(t, sha3.New256(), difficulty) {
		return errors.New("invalid ticket challenge")
	}

	// Contents are a properly encoded DID instance
	id, err := t.GetDID()
	if err != nil {
		return err
	}

	// Verify private keys are not included
	for _, k := range id.VerificationMethods() {
		if len(k.Private) != 0 {
			return errors.New("private keys included on the DID")
		}
	}

	// Retrieve DID key
	key := id.VerificationMethod(t.KeyId)
	if key == nil {
		return errors.New("the selected key is not available on the DID")
	}

	// Verify proof
	proof, err := t.GetProofLD()
	if err != nil {
		return err
	}
	data, err := id.Document(true).NormalizedLD()
	if err != nil {
		return err
	}
	if !key.VerifyProof(data, proof) {
		return errors.New("invalid proof")
	}

	// Get digest
	data, err = t.MarshalBinary()
	if err != nil {
		return errors.New("failed to re-encode ticket instance")
	}
	digest := sha3.New256()
	if _, err = digest.Write(data); err != nil {
		return err
	}

	// Verify signature
	if !key.Verify(digest.Sum(nil), t.Signature) {
		return errors.New("invalid ticket signature")
	}
	return nil
}
