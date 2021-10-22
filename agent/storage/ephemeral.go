package storage

import (
	"fmt"
	"sync"

	protov1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"github.com/pkg/errors"
	"go.bryk.io/pkg/did"
)

type record struct {
	id    *did.Identifier
	proof *did.ProofLD
}

// Ephemeral provides an in-memory store for development and testing.
type Ephemeral struct {
	entries map[string]*record
	mu      sync.Mutex
}

// Open is a no-op for the ephemeral store. As an example just setup
// internally used structures.
func (e *Ephemeral) Open(_ string) error {
	e.entries = make(map[string]*record)
	return nil
}

// Close will release used-memory.
func (e *Ephemeral) Close() error {
	e.mu.Lock()
	for k := range e.entries {
		delete(e.entries, k)
	}
	e.mu.Unlock()
	return nil
}

// Exists returns true if the provided DID instance is already available
// in the store.
func (e *Ephemeral) Exists(id *did.Identifier) bool {
	key := fmt.Sprintf("%s:%s", id.Method(), id.Subject())
	e.mu.Lock()
	_, ok := e.entries[key]
	e.mu.Unlock()
	return ok
}

// Get a previously stored DID instance.
func (e *Ephemeral) Get(req *protov1.QueryRequest) (*did.Identifier, *did.ProofLD, error) {
	key := fmt.Sprintf("%s:%s", req.Method, req.Subject)
	e.mu.Lock()
	r, ok := e.entries[key]
	e.mu.Unlock()
	if !ok {
		return nil, nil, errors.New("no information available")
	}
	return r.id, r.proof, nil
}

// Save will create or update an entry for the provided DID instance.
func (e *Ephemeral) Save(id *did.Identifier, proof *did.ProofLD) (string, error) {
	key := fmt.Sprintf("%s:%s", id.Method(), id.Subject())
	e.mu.Lock()
	e.entries[key] = &record{
		id:    id,
		proof: proof,
	}
	e.mu.Unlock()
	return "", nil
}

// Delete any existing record for the provided DID instance.
func (e *Ephemeral) Delete(id *did.Identifier) error {
	key := fmt.Sprintf("%s:%s", id.Method(), id.Subject())
	e.mu.Lock()
	delete(e.entries, key)
	e.mu.Unlock()
	return nil
}

// Description returns a brief summary for the storage instance.
func (e *Ephemeral) Description() string {
	return "ephemeral in-memory data store"
}
