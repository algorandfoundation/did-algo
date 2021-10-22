package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	ac "github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	xlog "go.bryk.io/pkg/log"
)

var walletRestoreCmd = &cobra.Command{
	Use:     "restore",
	Short:   "Restore a wallet using an existing mnemonic file",
	Aliases: []string{"recover"},
	Example: "algoid wallet restore [mnemonic-file]",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get parameters
		if len(args) != 1 {
			return errors.New("missing required parameters")
		}

		// Verify mnemonic file corresponds to a valid private key
		wp, err := ioutil.ReadFile(filepath.Clean(args[0]))
		if err != nil {
			return fmt.Errorf("failed to read mnemonic file: %w", err)
		}
		key, err := mnemonic.ToPrivateKey(string(wp))
		if err != nil {
			return err
		}
		account, err := ac.AccountFromPrivateKey(key)
		if err != nil {
			return err
		}

		// Get local store handler
		store, err := getClientStore()
		if err != nil {
			return err
		}

		// Ask wallet parameters
		name, err := readValue("enter a name for your wallet")
		if err != nil {
			return err
		}
		name = sanitize.Name(name)
		pass, err := readSecretValue("enter a secure passphrase for your wallet")
		if err != nil {
			return err
		}

		// Verify this won't replace an existing wallet
		if store.WalletExists(name) {
			return fmt.Errorf("a wallet named: '%s' already exists", name)
		}

		// Save wallet file
		if err := store.SaveWallet(name, string(wp), pass); err != nil {
			return err
		}
		log.WithFields(xlog.Fields{
			"name":    name,
			"address": account.Address.String(),
		}).Info("new wallet created")
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletRestoreCmd)
}
