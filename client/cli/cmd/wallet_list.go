package cmd

import (
	"github.com/spf13/cobra"
)

var walletListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List your existing ALGO wallet(s)",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getClientStore()
		if err != nil {
			return err
		}
		list := store.ListWallets()
		if len(list) == 0 {
			log.Warning("no wallets found")
			return nil
		}
		for _, el := range list {
			log.Infof("wallet found: %s", el)
		}
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletListCmd)
}
