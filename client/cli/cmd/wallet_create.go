package cmd

import (
	"errors"
	"fmt"

	ac "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	xlog "go.bryk.io/pkg/log"
)

var walletCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create a new (standalone) ALGO wallet",
	Example: "algoid wallet new [wallet-name]",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get parameters
		if len(args) != 1 {
			return errors.New("you must provide a name for your wallet")
		}
		name := sanitize.Name(args[0])
		wp, err := readSecretValue("enter a secure passphrase for your new wallet")
		if err != nil {
			return err
		}
		confirmation, err := readSecretValue("confirm your passphrase")
		if err != nil {
			return err
		}
		if wp != confirmation {
			return errors.New("passphrase confirmation failed")
		}

		// Get local store handler
		store, err := getClientStore()
		if err != nil {
			return err
		}

		// Verify this won't replace an existing wallet
		if store.WalletExists(name) {
			return fmt.Errorf("a wallet named: '%s' already exists", name)
		}

		// Create new account and securely store wallet contents
		account := ac.GenerateAccount()
		seed, err := mnemonic.FromPrivateKey(account.PrivateKey)
		if err != nil {
			return err
		}
		if err := store.SaveWallet(name, seed, wp); err != nil {
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
	walletCmd.AddCommand(walletCreateCmd)
}
