package storage

import (
	"context"
	"encoding/base64"
	"encoding/json"

	protov1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"github.com/pkg/errors"
	"go.bryk.io/pkg/did"
	"go.bryk.io/pkg/storage/orm"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Base64 encoding used
var b64 = base64.RawStdEncoding

// Data structure to store DID entries.
type identifierRecord struct {
	// DID method.
	Method string `bson:"method"`

	// DID subject.
	Subject string `bson:"subject"`

	// DID document.
	Document string `bson:"document"`

	// DID proof.
	Proof string `bson:"proof"`
}

func (ir *identifierRecord) decode() (*did.Identifier, *did.ProofLD, error) {
	d1, err := b64.DecodeString(ir.Document)
	if err != nil {
		return nil, nil, errors.New("invalid record contents")
	}
	d2, err := b64.DecodeString(ir.Proof)
	if err != nil {
		return nil, nil, errors.New("invalid record contents")
	}

	// Restore DID document
	doc := &did.Document{}
	if err = json.Unmarshal(d1, doc); err != nil {
		return nil, nil, errors.New("invalid record contents")
	}
	id, err := did.FromDocument(doc)
	if err != nil {
		return nil, nil, err
	}

	// Restore proof
	proof := &did.ProofLD{}
	if err = json.Unmarshal(d2, proof); err != nil {
		return nil, nil, errors.New("invalid record contents")
	}
	return id, proof, nil
}

func (ir *identifierRecord) encode(id *did.Identifier, proof *did.ProofLD) {
	data, _ := json.Marshal(id.Document(true))
	pp, _ := json.Marshal(proof)
	ir.Method = id.Method()
	ir.Subject = id.Subject()
	ir.Document = b64.EncodeToString(data)
	ir.Proof = b64.EncodeToString(pp)
}

// MongoStore provides a storage handler utilizing MongoDB as underlying
// database. The connection strings must be of the form "mongodb://...";
// for example: "mongodb://localhost:27017"
type MongoStore struct {
	op  *orm.Operator
	did *orm.Model
}

// Open establish the connection and database selection for the instance.
// Must be called before any further operations.
func (ms *MongoStore) Open(info string) error {
	var err error
	opts := options.Client()
	opts.ApplyURI(info)
	ms.op, err = orm.NewOperator("algoid", opts)
	if err != nil {
		return err
	}
	ms.did = ms.op.Model("identifiers")
	return err
}

// Close the client connection with the backend server.
func (ms *MongoStore) Close() error {
	return ms.op.Close(context.TODO())
}

// Exists returns true if the provided DID instance is already available
// in the store.
func (ms *MongoStore) Exists(id *did.Identifier) bool {
	n, _ := ms.did.Count(filter(id))
	return n > 0
}

// Get a previously stored DID instance.
func (ms *MongoStore) Get(req *protov1.QueryRequest) (*did.Identifier, *did.ProofLD, error) {
	// Run query
	var res identifierRecord
	filter := orm.Filter()
	filter["method"] = req.Method
	filter["subject"] = req.Subject
	if err := ms.did.First(filter, &res); err != nil {
		return nil, nil, errors.New("no information available")
	}

	// Decode result
	return res.decode()
}

// Save will create or update an entry for the provided DID instance.
func (ms *MongoStore) Save(id *did.Identifier, proof *did.ProofLD) (string, error) {
	// Record
	rec := new(identifierRecord)
	rec.encode(id, proof)

	// Run upsert operation
	return "", ms.did.Update(filter(id), rec, true)
}

// Delete any existing record for the provided DID instance.
func (ms *MongoStore) Delete(id *did.Identifier) error {
	return ms.did.Delete(filter(id))
}

// Description returns a brief summary for the storage instance.
func (ms *MongoStore) Description() string {
	return "MongoDB data store"
}

// Helper method to produce a selector from a DID instance.
func filter(id *did.Identifier) map[string]interface{} {
	filter := orm.Filter()
	filter["method"] = id.Method()
	filter["subject"] = id.Subject()
	return filter
}
