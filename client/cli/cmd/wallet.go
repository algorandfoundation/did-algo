package cmd

import (
	"fmt"

	ac "github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	"github.com/spf13/cobra"
	"go.bryk.io/pkg/cli"
)

type algoDestination struct {
	Address string `json:"address"`
	Network string `json:"network"`
	Asset   string `json:"asset"`
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Manage your ALGO wallet(s)",
}

func init() {
	rootCmd.AddCommand(walletCmd)
}

func readValue(prompt string) (string, error) {
	var dest string
	fmt.Printf("%s: ", prompt)
	_, err := fmt.Scanln(&dest)
	return dest, err
}

func readSecretValue(prompt string) (string, error) {
	v, err := cli.ReadSecure(fmt.Sprintf("%s: ", prompt))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(v), nil
}

func getWalletAddress(name, pass string) (string, error) {
	// Get local store handler
	store, err := getClientStore()
	if err != nil {
		return "", err
	}

	// Decrypt wallet
	seed, err := store.OpenWallet(name, pass)
	if err != nil {
		return "", err
	}

	// Restore account handler
	key, err := mnemonic.ToPrivateKey(seed)
	if err != nil {
		return "", err
	}
	account, err := ac.AccountFromPrivateKey(key)
	if err != nil {
		return "", err
	}

	// Return account address
	return account.Address.String(), nil
}
