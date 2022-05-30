package cmd

import (
	"context"
	"errors"
	"fmt"

	ac "github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
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

		// Get client connection
		conn, err := getClientConnection()
		if err != nil {
			return fmt.Errorf("failed to establish connection: %s", err)
		}
		defer func() {
			_ = conn.Close()
		}()
		cl := protoV1.NewAgentAPIClient(conn)

		// Get account info
		info, err := cl.AccountInformation(context.TODO(), &protoV1.AccountInformationRequest{
			Address:  account.Address.String(),
			Protocol: "algorand",
			Network:  "testnet",
		})
		if err != nil {
			return err
		}

		// Print results
		fmt.Printf("address: %s\n", account.Address.String())
		fmt.Printf("status: %s\n", info.Status)
		fmt.Printf("current balance: %d\n", info.Balance)
		fmt.Printf("pending rewards: %d\n", info.PendingRewards)
		fmt.Printf("total rewards: %d\n", info.TotalRewards)
		for _, pt := range info.PendingTransactions {
			fmt.Printf("pending transaction. amount: %d, to: %s\n", pt.Amount, pt.Receiver)
		}
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletInfoCmd)
}
