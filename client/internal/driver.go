package internal

import (
	"context"
	"encoding/json"

	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"go.bryk.io/pkg/did"
	"go.bryk.io/pkg/did/resolver"
	"go.bryk.io/pkg/errors"
)

// provider required for the resolver provider.
type provider struct {
	client protoV1.AgentAPIClient
}

func (p *provider) Read(id string) (*did.Document, *did.DocumentMetadata, error) {
	ID, err := did.Parse(id)
	if err != nil {
		return nil, nil, errors.New(resolver.ErrInvalidDID)
	}
	req := &protoV1.QueryRequest{
		Method:  ID.Method(),
		Subject: ID.Subject(),
	}
	res, err := p.client.Query(context.Background(), req)
	if err != nil {
		return nil, nil, errors.New(resolver.ErrNotFound)
	}
	doc := new(did.Document)
	if err = json.Unmarshal(res.Document, doc); err != nil {
		return nil, nil, errors.New(resolver.ErrInvalidDocument)
	}
	md := new(did.DocumentMetadata)
	if err = json.Unmarshal(res.DocumentMetadata, md); err != nil {
		return nil, nil, errors.New(resolver.ErrInvalidDocument)
	}
	return doc, md, nil
}
