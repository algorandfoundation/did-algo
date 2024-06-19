package cmd

import (
	"errors"
	"fmt"

	ac "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var deployContractCmd = &cobra.Command{
	Use:     "deploy",
	Aliases: []string{"deploy-contract"},
	Short:   "Deploy the DIDAlgoStorage smart contract",
	Example: "algoid deploy [wallet-name] [network]",
	RunE:    runDeployContractCmd,
}

func init() {
	rootCmd.AddCommand(deployContractCmd)
}

func runDeployContractCmd(_ *cobra.Command, args []string) error {
	// Get parameters
	if len(args) != 2 {
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

	network := args[1]
	// Deploy contract
	appID, err := cl.Networks[network].DeployContract(&account)
	if err != nil {
		return err
	}
	log.WithField("app_id", appID).Info(fmt.Sprintf("storage contract deployed successfully to %s", network))
	return nil
}
