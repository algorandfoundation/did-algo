package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/algorandfoundation/did-algo/client/store"
	"github.com/algorandfoundation/did-algo/info"
	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"go.bryk.io/pkg/did"
	xlog "go.bryk.io/pkg/log"
	"go.bryk.io/pkg/net/rpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Provider is responsible for handling features required
// by the local API server.
type Provider struct {
	st     *store.LocalStore
	log    xlog.Logger
	conf   *internal.ClientSettings
	conn   *grpc.ClientConn
	client protoV1.AgentAPIClient
}

// LocalEntry represents a DID instance stored in the local
// store using a simplified format suitable for the UI.
type LocalEntry struct {
	// DID local reference name
	Name string `json:"name"`

	// DID identifier
	DID string `json:"did"`

	// ALGO addresses linked to the identifier
	Addresses []addressEntry `json:"addresses"`

	// Whether the DID remains active
	Active bool `json:"active"`

	// Date when the DID was last sync to the network
	LastSync string `json:"last_sync"`

	// Raw DID document
	Document *did.Document `json:"document"`
}

// Ready returns true if the network is available.
func (p *Provider) Ready() bool {
	res, err := p.client.Ping(context.Background(), &emptypb.Empty{})
	if err != nil {
		return false
	}
	return res.Ok
}

// List all DID instances in the local store.
func (p *Provider) List() []*LocalEntry {
	// Get list of entries
	result := []*LocalEntry{}
	for k, id := range p.st.List() {
		entry := &LocalEntry{
			Name:      k,
			DID:       id.DID(),
			Document:  id.Document(true),
			Active:    !id.GetMetadata().Deactivated,
			Addresses: getAlgoAddress(id),
			LastSync:  id.GetMetadata().Updated,
		}
		result = append(result, entry)
	}
	return result
}

// Register a new DID instance in the local store.
func (p *Provider) Register(name string, passphrase string) error {
	// Check for duplicates
	dup, _ := p.st.Get(name)
	if dup != nil {
		return fmt.Errorf("there's already a DID with reference name: %s", name)
	}

	// Generate master key from available secret
	masterKey, err := keyFromMaterial([]byte(passphrase))
	if err != nil {
		return err
	}
	defer masterKey.Destroy()
	pk := make([]byte, 64)
	copy(pk, masterKey.PrivateKey())

	p.log.Info("generating new identifier")
	id, err := did.NewIdentifierWithMode("algo", "", did.ModeUUID)
	if err != nil {
		return err
	}
	p.log.Debug("adding master key")
	if err = id.AddVerificationMethod("master", pk, did.KeyTypeEd); err != nil {
		return err
	}
	p.log.Debug("setting master key as authentication mechanism")
	if err = id.AddVerificationRelationship(id.GetReference("master"), did.AuthenticationVM); err != nil {
		return err
	}

	// Save instance in the store
	p.log.Info("adding entry to local store")
	return p.st.Save(name, id)
}

// Sync a DID instance with the network.
func (p *Provider) Sync(name string) error {
	id, err := p.st.Get(name)
	if err != nil {
		return fmt.Errorf("no available record under the provided reference name: %s", name)
	}

	// Get selected key for the sync operation
	key, err := getSyncKey(id)
	if err != nil {
		return err
	}
	p.log.Debugf("key selected for the operation: %s", key.ID)

	// Generate request ticket
	p.log.Infof("publishing: %s", name)
	ticket, err := getRequestTicket(id, key)
	if err != nil {
		return err
	}
	req := &protoV1.ProcessRequest{Ticket: ticket}

	// Submit request
	p.log.Info("submitting request to the network")
	res, err := p.client.Process(context.Background(), req)
	if err != nil {
		return fmt.Errorf("network return an error: %w", err)
	}
	p.log.Debugf("request status: %v", res.Ok)
	if res.Identifier != "" {
		p.log.Info("identifier: ", res.Identifier)
	}
	if !res.Ok {
		return nil
	}

	// Update local record if sync was successful
	return p.st.Update(name, id)
}

// Update and sync a DID instance with the network.
func (p *Provider) Update(req *updateRequest) error {
	// retrieve local record
	id, err := p.st.Get(req.Name)
	if err != nil {
		return fmt.Errorf("no available record under the provided reference name: %s", req.Name)
	}

	// get service entry
	svc := id.Service("algo-connect")
	if svc == nil {
		svc = newServiceEntry()
	}

	// replace addresses
	var addresses []algoDestination
	ext := did.Extension{
		ID:      "algo-address",
		Version: "0.1.0",
	}
	for _, entry := range req.Addresses {
		if entry.Enabled {
			addresses = append(addresses, algoDestination{
				Address: entry.Address,
				Network: strings.ToLower(entry.Network),
				Asset:   "ALGO",
			})
		}
	}
	ext.Data = addresses
	svc.AddExtension(ext)

	// update service entry
	_ = id.RemoveService("algo-connect")
	if len(addresses) > 0 {
		if err := id.AddService(svc); err != nil {
			return err
		}
		id.RegisterContext("https://did-ns.aidtech.network/v1")
	}

	// update local record
	if err = p.st.Update(req.Name, id); err != nil {
		return err
	}

	// sync with the network in the background
	go func() {
		_ = p.Sync(req.Name)
	}()
	return nil
}

// Get client connection.
func (p *Provider) connect() error {
	p.log.Infof("establishing connection to network agent: %s", p.conf.Node)
	opts := []rpc.ClientOption{
		rpc.WaitForReady(),
		rpc.WithUserAgent(fmt.Sprintf("algoid-client/%s", info.CoreVersion)),
		rpc.WithTimeout(time.Duration(p.conf.Timeout) * time.Second),
	}
	if p.conf.Insecure {
		p.log.Warning("using an insecure connection")
	} else {
		opts = append(opts, rpc.WithClientTLS(rpc.ClientTLSConfig{IncludeSystemCAs: true}))
	}
	if p.conf.Override != "" {
		p.log.WithField("override", p.conf.Override).Warning("using server name override")
		opts = append(opts, rpc.WithServerNameOverride(p.conf.Override))
	}
	conn, err := rpc.NewClientConnection(p.conf.Node, opts...)
	if err != nil {
		return err
	}
	p.client = protoV1.NewAgentAPIClient(conn)
	return nil
}

// Close client connection and free resources.
func (p *Provider) close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

type algoDestination struct {
	Address string `json:"address"`
	Network string `json:"network"`
	Asset   string `json:"asset"`
}

type addressEntry struct {
	Address string `json:"address"`
	Network string `json:"network"`
	Enabled bool   `json:"enabled"`
}

type updateRequest struct {
	Name      string         `json:"name"`
	DID       string         `json:"did"`
	Addresses []addressEntry `json:"addresses"`
}
