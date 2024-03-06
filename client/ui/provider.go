package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	ac "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/algorandfoundation/did-algo/client/store"
	"go.bryk.io/pkg/did"
	xlog "go.bryk.io/pkg/log"
)

// Provider is responsible for handling features required
// by the local API server.
type Provider struct {
	st     *store.LocalStore
	log    xlog.Logger
	client *internal.AlgoClient
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
	return p.client.Ready()
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

	// Create wallet
	account := ac.GenerateAccount()
	seed, err := mnemonic.FromPrivateKey(account.PrivateKey)
	if err != nil {
		return err
	}
	if err := p.st.SaveWallet(name, seed, passphrase); err != nil {
		return err
	}

	// Generate base identifier instance
	subject := fmt.Sprintf("%x-%d", account.PublicKey, p.client.StorageAppID())
	method := "algo"
	p.log.WithFields(xlog.Fields{
		"subject": subject,
		"method":  method,
	}).Info("generating new identifier")
	id, err := did.NewIdentifier(method, subject)
	if err != nil {
		return err
	}
	if err = id.AddVerificationMethod("master", account.PrivateKey, did.KeyTypeEd); err != nil {
		return err
	}
	if err = id.AddVerificationRelationship(id.GetReference("master"), did.AuthenticationVM); err != nil {
		return err
	}

	// Save instance in the store
	p.log.Info("adding entry to local store")
	return p.st.Save(name, id)
}

// Sync a DID instance with the network.
func (p *Provider) Sync(name string, passphrase string) error {
	id, err := p.st.Get(name)
	if err != nil {
		return fmt.Errorf("no available record under the provided reference name: %s", name)
	}

	// Decrypt wallet
	seed, err := p.st.OpenWallet(name, passphrase)
	if err != nil {
		return err
	}

	// Restore account handler
	key, err := mnemonic.ToPrivateKey(seed)
	if err != nil {
		return err
	}
	account, err := ac.AccountFromPrivateKey(key)
	if err != nil {
		return err
	}

	// Submit request
	p.log.Info("submitting request to the network")
	if err := p.client.PublishDID(id, &account); err != nil {
		return fmt.Errorf("network return an error: %w", err)
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

	if len(req.Addresses) != 0 { // nolint: nestif
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
			id.RegisterContext("https://did.algorand.foundation/v1")
		}

		// update local record
		if err = p.st.Update(req.Name, id); err != nil {
			return err
		}
	}

	// sync with the network in the background
	go func() {
		_ = p.Sync(req.Name, req.Passphrase)
	}()
	return nil
}

// ServerHandler returns an HTTP handler that can be used to exposed
// the provider instance through an HTTP server.
func (p *Provider) ServerHandler() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/ready", p.readyHandlerFunc)
	router.HandleFunc("/list", p.listHandlerFunc)
	router.HandleFunc("/register", p.registerHandlerFunc)
	router.HandleFunc("/update", p.updateHandlerFunc)
	return router
}

// Close client connection and free resources.
func (p *Provider) close() error {
	return nil
}

// HTTP handler for the `GET /ready` endpoint.
func (p *Provider) readyHandlerFunc(w http.ResponseWriter, _ *http.Request) {
	if !p.Ready() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	_, _ = w.Write([]byte("ok"))
}

// HTTP handler for the `GET /list` endpoint.
func (p *Provider) listHandlerFunc(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(p.List())
}

// HTTP handler for the `POST /register` endpoint.
func (p *Provider) registerHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()
	params := map[string]string{}
	if err = json.Unmarshal(body, &params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	name, ok := params["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	passphrase, ok := params["recovery_key"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = p.Register(name, passphrase); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte("ok"))
}

// HTTP handler for the `POST /update` endpoint.
func (p *Provider) updateHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()
	req := new(updateRequest)
	if err = json.Unmarshal(body, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = p.Update(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write([]byte("ok"))
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
	Name       string         `json:"name"`
	DID        string         `json:"did"`
	Passphrase string         `json:"passphrase"`
	Addresses  []addressEntry `json:"addresses"`
}
