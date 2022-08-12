package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var walletExportCmd = &cobra.Command{
	Use:     "export",
	Aliases: []string{"save"},
	Example: "algoid wallet export [wallet-name]",
	Short:   "Export wallet's master derivation key",
	Long: `Export wallet's master derivation key.

A master derivation key can be exported as a mnemonic in
order to allow a user to securely restore a wallet an its
cryptographic material to prevent loosing access to the
assets protected by it.

The mnemonic will be automatically saved to a text file
with the name "[wallet-name]-mnemonic.txt".

Keep in mind that misplacing or sharing the mnemonic can
result in catastrophic security issues and permanent loss
of funds.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get parameters
		if len(args) != 1 {
			return errors.New("missing required parameters")
		}
		name := sanitize.Name(args[0])
		wp, err := readSecretValue("enter wallet's passphrase")
		if err != nil {
			return err
		}

		// Get local store handler
		store, err := getClientStore()
		if err != nil {
			return err
		}

		// Decrypt wallet
		seed, err := store.OpenWallet(name, wp)
		if err != nil {
			return err
		}

		// Save export file
		fileName := fmt.Sprintf("%s-mnemonic.txt", name)
		err = os.WriteFile(fileName, []byte(seed), 0400)
		if err != nil {
			return err
		}
		log.Infof("data stored in file: %s", fileName)
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletExportCmd)
}
