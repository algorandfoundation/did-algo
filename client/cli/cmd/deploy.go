package cmd

import (
	"errors"

	ac "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var deployContractCmd = &cobra.Command{
	Use:     "deploy",
	Aliases: []string{"deploy-contract"},
	Short:   "Deploy the AlgoDID storage smart contract",
	Example: "algoid deploy [wallet-name]",
	RunE:    runDeployContractCmd,
}

func init() {
	rootCmd.AddCommand(deployContractCmd)
}

func runDeployContractCmd(_ *cobra.Command, args []string) error {
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

	// Deploy contract
	appID, err := cl.DeployContract(&account)
	if err != nil {
		return err
	}
	log.WithField("app_id", appID).Info("storage contract deployed successfully")
	return nil
}
