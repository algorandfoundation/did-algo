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
	"go.bryk.io/pkg/did"
	xlog "go.bryk.io/pkg/log"
)

var registerCmd = &cobra.Command{
	Use:     "register",
	Short:   "Creates a new DID locally",
	Example: "algoid register [wallet-name]",
	Aliases: []string{"create", "new"},
	RunE:    runRegisterCmd,
}

func init() {
	params := []cli.Param{}
	if err := cli.SetupCommandParams(registerCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(registerCmd)
}

func runRegisterCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("missing required parameters")
	}
	name := sanitize.Name(args[0])
	wp, err := readSecretValue("enter wallet's passphrase")
	if err != nil {
		return err
	}

	// Get local store handler
	st, err := getClientStore()
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

	// Get storage application identifier
	appID, err := getStorageAppID()
	if err != nil {
		return err
	}

	// Check for duplicates
	dup, _ := st.Get(name)
	if dup != nil {
		return fmt.Errorf("there's already a DID with reference name: %s", name)
	}

	// TODO: pass this in from cli
	network := "testnet"
	// Generate base identifier instance
	subject := fmt.Sprintf("%s-%x-%d", network, account.PublicKey, appID)
	method := "algo"
	log.WithFields(xlog.Fields{
		"subject": subject,
		"method":  method,
	}).Info("generating new identifier")
	id, err := did.NewIdentifier(method, subject)
	if err != nil {
		return err
	}
	log.Debug("adding master key")
	if err = id.AddVerificationMethod("master", key, did.KeyTypeEd); err != nil {
		return err
	}
	log.Debug("setting master key as authentication mechanism")
	if err = id.AddVerificationRelationship(id.GetReference("master"), did.AuthenticationVM); err != nil {
		return err
	}

	// Save instance in the store
	log.Info("adding entry to local store")
	return st.Save(name, id)
}
