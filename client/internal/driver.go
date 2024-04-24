package internal

import (
	"errors"

	"go.bryk.io/pkg/did"
	"go.bryk.io/pkg/did/resolver"
)

// Read a DID document from the Algorand network. The method complies
// with the `resolver.Provider` interface.
func (c *AlgoDIDClient) Read(id string) (*did.Document, *did.DocumentMetadata, error) {
	if _, err := did.Parse(id); err != nil {
		return nil, nil, errors.New(resolver.ErrInvalidDID)
	}
	doc, err := c.Resolve(id)
	if err != nil {
		return nil, nil, errors.New(resolver.ErrNotFound)
	}
	md := new(did.DocumentMetadata)
	md.Deactivated = false
	return doc, md, nil
}
