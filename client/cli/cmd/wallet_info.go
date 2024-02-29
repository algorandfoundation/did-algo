package cmd

import (
	"errors"
	"fmt"

	ac "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var walletInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Get account information",
	Example: "algoid wallet info [wallet-name]",
	Aliases: []string{"details", "inspect", "more"},
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

		// Restore account handler
		key, err := mnemonic.ToPrivateKey(seed)
		if err != nil {
			return err
		}
		account, err := ac.AccountFromPrivateKey(key)
		if err != nil {
			return err
		}

		// Get network client
		cl, err := getAlgoClient()
		if err != nil {
			return err
		}

		// Get account info
		info, err := cl.AccountInformation(account.Address.String())
		if err != nil {
			return err
		}

		// Print results
		fmt.Printf("address: %s\n", account.Address.String())
		fmt.Printf("public key: %x\n", account.PublicKey)
		fmt.Printf("status: %s\n", info.Status)
		fmt.Printf("round: %d\n", info.Round)
		fmt.Printf("current balance: %d\n", info.Amount)
		fmt.Printf("pending rewards: %d\n", info.PendingRewards)
		fmt.Printf("total rewards: %d\n", info.Rewards)
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletInfoCmd)
}
