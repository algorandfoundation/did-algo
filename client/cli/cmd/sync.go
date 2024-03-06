package cmd

import (
	"errors"
	"fmt"

	ac "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
)

var syncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Publish a DID instance to the processing network",
	Example: "algoid sync [DID name]",
	Aliases: []string{"publish", "update", "upload", "push"},
	RunE:    runSyncCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "delete",
			Usage:     "mark the DID instance as deleted",
			FlagKey:   "sync.delete",
			ByDefault: false,
			Short:     "d",
		},
	}
	if err := cli.SetupCommandParams(syncCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(syncCmd)
}

func runSyncCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("missing required parameters")
	}

	// Get store handler
	st, err := getClientStore()
	if err != nil {
		return err
	}

	// Retrieve identifier
	name := sanitize.Name(args[0])
	id, err := st.Get(name)
	if err != nil {
		return fmt.Errorf("no available record under reference name: %s", name)
	}

	// Get wallet account
	wp, err := readSecretValue("enter wallet's passphrase")
	if err != nil {
		return err
	}

	// Decrypt wallet
	seed, err := st.OpenWallet(name, wp)
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

	// Submit request
	log.Info("submitting request to the network")
	if viper.GetBool("sync.delete") {
		log.Infof("deleting: %s", name)
		if err = cl.DeleteDID(id, &account); err != nil {
			return err
		}
		log.Info("DID instance deleted")
		return nil
	}
	log.Infof("publishing: %s", name)
	if err := cl.PublishDID(id, &account); err != nil {
		return err
	}
	log.Info("DID instance published")
	return nil
}
