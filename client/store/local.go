package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.bryk.io/pkg/crypto/tred"
	"go.bryk.io/pkg/did"
)

// Local storage version.
const currentVersion = "0.1.0"

// LocalStore provides a filesystem-backed store.
type LocalStore struct {
	home string
}

// IdentifierRecord holds the identifier data, Document and DocumentMetadata,
// and is used to store the identifier locally.
type IdentifierRecord struct {
	Version  string                `json:"version,omitempty"`
	Document *did.Document         `json:"document"`
	Metadata *did.DocumentMetadata `json:"metadata,omitempty"`
	Proof    *did.ProofLD          `json:"proof,omitempty"`
}

// NewLocalStore returns a local store handler. If the specified
// 'home' directory doesn't exist it will be created.
func NewLocalStore(home string) (*LocalStore, error) {
	h := filepath.Clean(home)
	if !dirExist(h) {
		if err := os.Mkdir(h, 0700); err != nil {
			return nil, fmt.Errorf("failed to create new home directory: %w", err)
		}
	}
	if !dirExist(filepath.Join(h, "wallets")) {
		if err := os.Mkdir(filepath.Join(h, "wallets"), 0700); err != nil {
			return nil, fmt.Errorf("failed to create new wallets directory: %w", err)
		}
	}
	return &LocalStore{home: h}, nil
}

// Save add a new entry to the store.
func (ls *LocalStore) Save(name string, id *did.Identifier) error {
	data, err := json.Marshal(&IdentifierRecord{
		Version:  currentVersion,
		Document: id.Document(false),
		Metadata: id.GetMetadata(),
	})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ls.home, name), data, 0600)
}

// Get an existing entry based on its reference name.
func (ls *LocalStore) Get(name string) (*did.Identifier, error) {
	data, err := os.ReadFile(filepath.Clean(filepath.Join(ls.home, name)))
	if err != nil {
		return nil, err
	}
	ir := new(IdentifierRecord)
	if err := json.Unmarshal(data, ir); err != nil {
		return nil, err
	}

	id, err := did.FromDocument(ir.Document)
	if err != nil {
		return nil, err
	}

	if ir.Metadata != nil {
		if err := id.AddMetadata(ir.Metadata); err != nil {
			return nil, err
		}
	}

	return id, nil
}

// List currently registered entries.
func (ls *LocalStore) List() map[string]*did.Identifier {
	// nolint: prealloc
	var list = make(map[string]*did.Identifier)
	_ = filepath.Walk(ls.home, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		id, err := ls.Get(info.Name())
		if err == nil {
			list[info.Name()] = id
		}
		return nil
	})
	return list
}

// Update the contents of an existing entry.
func (ls *LocalStore) Update(name string, id *did.Identifier) error {
	return ls.Save(name, id)
}

// Delete a previously stored entry.
func (ls *LocalStore) Delete(name string) error {
	return os.Remove(filepath.Join(ls.home, name))
}

// WalletExists returns `true` if a wallet with the provided name
// already exists.
func (ls *LocalStore) WalletExists(name string) bool {
	wf := filepath.Join(ls.home, "wallets", fmt.Sprintf("%s.wal", name))
	info, err := os.Stat(wf)
	return err == nil && !info.IsDir()
}

// SaveWallet will securely store the wallet contents on the local FS.
// The `passphrase` value will be used to encrypt the `mnemonic` value
// before storing it.
func (ls *LocalStore) SaveWallet(name, mnemonic, passphrase string) error {
	// Create a tred worker instance
	tw, err := tredWorker([]byte(passphrase))
	if err != nil {
		return err
	}

	// Encrypt wallet contents and save
	enc := bytes.NewBuffer(nil)
	if _, err := tw.Encrypt(bytes.NewReader([]byte(mnemonic)), enc); err != nil {
		return err
	}
	wf := filepath.Join("wallets", fmt.Sprintf("%s.wal", name))
	return os.WriteFile(filepath.Join(ls.home, wf), enc.Bytes(), 0400)
}

// OpenWallet locates the wallet file and decrypts it using `passphrase`;
// the `mnemonic` value returned must be treated with extreme caution.
func (ls *LocalStore) OpenWallet(name, passphrase string) (string, error) {
	// Open encrypted wallet file
	wf := filepath.Join("wallets", fmt.Sprintf("%s.wal", name))
	contents, err := os.ReadFile(filepath.Clean(filepath.Join(ls.home, wf)))
	if err != nil {
		return "", err
	}

	// Create a tred worker instance
	tw, err := tredWorker([]byte(passphrase))
	if err != nil {
		return "", err
	}

	// Decrypt wallet file
	dec := bytes.NewBuffer(nil)
	if _, err := tw.Decrypt(bytes.NewReader(contents), dec); err != nil {
		return "", err
	}
	return dec.String(), nil
}

// ListWallets returns a list of the names of all wallets locally stored.
func (ls *LocalStore) ListWallets() (list []string) {
	_ = filepath.Walk(filepath.Join(ls.home, "wallets"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".wal") {
			list = append(list, strings.TrimSuffix(info.Name(), ".wal"))
		}
		return nil
	})
	return
}

// RenameWallet will securely adjust the alias associated with a locally
// store wallet.
func (ls *LocalStore) RenameWallet(old, nw string) error {
	wf := filepath.Join(ls.home, "wallets", fmt.Sprintf("%s.wal", old))
	nf := filepath.Join(ls.home, "wallets", fmt.Sprintf("%s.wal", nw))
	return os.Rename(wf, nf)
}

// DeleteWallet will permanently delete the file handler for an existing
// wallet. This cannot be undone. Use with EXTREME care.
func (ls *LocalStore) DeleteWallet(name string) error {
	wf := filepath.Join(ls.home, "wallets", fmt.Sprintf("%s.wal", name))
	return os.Remove(wf)
}

// Verify the provided path exists and is a directory.
func dirExist(name string) bool {
	info, err := os.Stat(name)
	return err == nil && info.IsDir()
}

// Create a new TRED worker with the provided secret key. The worker
// instance can be used to secure data at-rest.
func tredWorker(key []byte) (*tred.Worker, error) {
	conf, err := tred.DefaultConfig(key)
	if err != nil {
		return nil, err
	}
	return tred.NewWorker(conf)
}
