package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	ipfs "github.com/ipfs/go-ipfs-api"
	"go.bryk.io/pkg/did"
)

// Default DNS link used for the did-algo index.
const indexDNSLink = "did-algo.aidtech.network"

// IPFS provides an integration with the "InterPlanetary File System",
// a decentralized global storage mechanism.
type IPFS struct {
	cl    *ipfs.Shell
	addr  string
	index string
}

// Open a connection with provided IPFS daemon instance.
func (c *IPFS) Open(addr string) error {
	sh := ipfs.NewShell(addr)
	if _, _, err := sh.Version(); err != nil {
		return fmt.Errorf("failed to connect to IPFS server: %w", err)
	}
	c.cl = sh
	c.cl.SetTimeout(time.Duration(5) * time.Second)
	c.addr = addr
	return nil
}

// Close the storage instance and free any resources in use.
func (c *IPFS) Close() error {
	return nil
}

// Description of the storage instance.
func (c *IPFS) Description() string {
	return fmt.Sprintf("IPFS data store [%s]", c.addr)
}

// Exists will check if a record exists for the specified DID.
func (c *IPFS) Exists(id *did.Identifier) bool {
	return c.existsInIndex(id.Subject())
}

// Get a previously stored DID instance.
func (c *IPFS) Get(req *protoV1.QueryRequest) (*did.Identifier, *did.ProofLD, error) {
	// Get CID from index
	cid := c.getIndexEntry(req.Subject)
	if cid == "" {
		return nil, nil, errors.New("no details for the requested DID")
	}

	// Read entry contents
	ptr, err := c.cl.Cat(cid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read record from IPFS: %w", err)
	}
	contents, err := io.ReadAll(ptr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read record from IPFS: %w", err)
	}

	// Decode contents
	dec := map[string]interface{}{}
	if err = json.Unmarshal(contents, &dec); err != nil {
		return nil, nil, fmt.Errorf("failed to decode record from IPFS: %w", err)
	}
	if _, ok := dec["document"]; !ok {
		return nil, nil, errors.New("invalid record contents, missing 'document'")
	}
	if _, ok := dec["proof"]; !ok {
		return nil, nil, errors.New("invalid record contents, missing 'proof'")
	}

	// Restore DID document
	doc := &did.Document{}
	docData, _ := json.Marshal(dec["document"])
	if err = json.Unmarshal(docData, doc); err != nil {
		return nil, nil, errors.New("invalid record contents on 'document'")
	}
	id, err := did.FromDocument(doc)
	if err != nil {
		return nil, nil, err
	}

	// Restore proof
	proof := &did.ProofLD{}
	proofData, _ := json.Marshal(dec["proof"])
	if err = json.Unmarshal(proofData, proof); err != nil {
		return nil, nil, errors.New("invalid record contents on 'proof'")
	}

	if _, ok := dec["metadata"]; ok {
		metadata := &did.DocumentMetadata{}
		metadataData, _ := json.Marshal(dec["metadata"])
		if err = json.Unmarshal(metadataData, metadata); err != nil {
			return nil, nil, errors.New("invalid record contents on 'metadata'")
		}
		if err := id.AddMetadata(metadata); err != nil {
			return nil, nil, err
		}
	}

	// Return final results
	return id, proof, nil
}

// Save the record for the given DID instance.
func (c *IPFS) Save(id *did.Identifier, proof *did.ProofLD) (string, error) {
	// Record to be stored on IPFS include the DID document and
	// its cryptographic proof for verification
	record := map[string]interface{}{
		"document": id.Document(true),
		"proof":    proof,
		"metadata": id.GetMetadata(),
	}
	data, err := json.Marshal(record)
	if err != nil {
		return "", err
	}

	// Store on IPFS using CID v1
	opts := []ipfs.AddOpts{
		ipfs.CidVersion(1),
		ipfs.Pin(true),
	}
	cid, err := c.cl.Add(bytes.NewReader(data), opts...)
	if err != nil {
		return "", err
	}

	// Create entry index for subject / CID. This will be used
	// when resolving the DID.
	return "/ipfs/" + cid, c.updateIndex(id.Subject(), cid)
}

// Delete any existing records for the given DID instance.
func (c *IPFS) Delete(_ *did.Identifier) error {
	return errors.New("IPFS entries cannot be removed")
}

func (c *IPFS) updateIndex(subject, cid string) (err error) {
	// Get index handler
	index, err := c.getIndexHandler()
	if err != nil {
		return fmt.Errorf("failed to open index handler: %w", err)
	}
	defer func() {
		_ = index.Close()
	}()

	// Load index contents
	var line string
	list := map[string]string{}
	scanner := bufio.NewScanner(index)
	for scanner.Scan() {
		line = scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		segs := strings.Split(line, " ")
		list[segs[0]] = segs[1]
	}

	// Add new entry and save new index contents
	list[subject] = cid
	contents := bytes.NewBuffer(nil)
	for subject, cid := range list {
		contents.WriteString(fmt.Sprintf("%s %s\n", subject, cid))
	}

	// Update index contents
	indexCID, err := c.cl.Add(contents, ipfs.CidVersion(1), ipfs.Pin(true))
	if err != nil {
		return fmt.Errorf("failed to update index contents: %w", err)
	}

	// Update index IPNS record (async)
	go func() {
		defer c.cl.SetTimeout(time.Duration(5) * time.Second)
		c.cl.SetTimeout(time.Duration(120) * time.Second)
		_, err := c.cl.PublishWithDetails(indexCID, "", 0, time.Duration(0), false)
		if err != nil {
			fmt.Printf("publish error: %s", err)
		}

		// Update agent's index handler
		if c.index, err = c.cl.Resolve(indexDNSLink); err != nil {
			fmt.Printf("update index error: %s", err)
		}
	}()
	return nil
}

func (c *IPFS) existsInIndex(subject string) bool {
	// Get index handler
	index, err := c.getIndexHandler()
	if err != nil {
		return false
	}
	defer func() {
		_ = index.Close()
	}()

	// Perform lookup operation
	var line string
	result := false
	scanner := bufio.NewScanner(index)
	for scanner.Scan() {
		line = scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Split(line, " ")[0] == subject {
			result = true
			break
		}
	}
	return result
}

func (c *IPFS) getIndexEntry(subject string) string {
	// Get index handler
	index, err := c.getIndexHandler()
	if err != nil {
		return ""
	}
	defer func() {
		_ = index.Close()
	}()

	// Perform lookup operation
	var line string
	scanner := bufio.NewScanner(index)
	for scanner.Scan() {
		line = scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if segs := strings.Split(line, " "); segs[0] == subject {
			if len(segs) != 2 {
				return ""
			}
			return segs[1]
		}
	}
	return ""
}

func (c *IPFS) getIndexHandler() (io.ReadCloser, error) {
	if c.index == "" {
		var err error
		c.index, err = c.cl.Resolve(indexDNSLink)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve index entry: %w", err)
		}
	}
	return c.cl.Cat(c.index)
}
