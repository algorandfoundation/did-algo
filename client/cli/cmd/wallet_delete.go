package cmd

import (
	"errors"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var walletDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "rm"},
	Example: "algoid wallet delete [wallet-name]",
	Short:   "Permanently delete an ALGO wallet",
	Long: `Permanently delete an ALGO wallet

This operation cannot be undone and might result in permanent
loss of funds. Use with EXTREME care.

If you are trying to move a wallet handler to a different system
be sure to first 'export' it to generate a backup file containing
the mnemonic that can be used later to 'restore' the wallet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("you must specify the wallet name")
		}

		// Get parameters
		name := sanitize.Name(args[0])
		wp, err := readSecretValue("enter wallet's passphrase")
		if err != nil {
			return err
		}

		// Verify wallet parameters
		store, err := getClientStore()
		if err != nil {
			return err
		}
		_, err = store.OpenWallet(name, wp)
		if err != nil {
			return err
		}

		// Get user confirmation
		confirmation, err := readValue("this action cannot be undone, are you sure (y/N)")
		if err != nil {
			return err
		}
		if confirmation != "y" {
			return errors.New("invalid confirmation value, canceling operation")
		}

		// Run delete operation
		if err = store.DeleteWallet(name); err != nil {
			return err
		}
		log.Infof("wallet '%s' was permanently deleted", name)
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletDeleteCmd)
}
