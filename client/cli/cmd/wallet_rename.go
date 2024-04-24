package cmd

import (
	"errors"
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var walletRenameCmd = &cobra.Command{
	Use:     "rename",
	Short:   "Rename an existing ALGO wallet",
	Example: "algoid wallet rename [current-name] [new-name]",
	Aliases: []string{"mv"},
	RunE: func(_ *cobra.Command, args []string) error {
		// Get parameters
		if len(args) != 2 {
			return errors.New("missing required parameters")
		}
		currentName := sanitize.Name(args[0])
		newName := sanitize.Name(args[1])

		// Get local store handler
		store, err := getClientStore()
		if err != nil {
			return err
		}

		// Verify this won't replace an existing wallet
		if store.WalletExists(newName) {
			return fmt.Errorf("a wallet named: '%s' already exists", newName)
		}

		// Rename wallet
		if err := store.RenameWallet(currentName, newName); err != nil {
			return err
		}
		log.Info("wallet renamed successfully")
		return nil
	},
}

func init() {
	// walletCmd.AddCommand(walletRenameCmd)
}
