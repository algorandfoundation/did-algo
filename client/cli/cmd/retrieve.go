package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/algorandfoundation/did-algo/client/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/did"
)

var retrieveCmd = &cobra.Command{
	Use:     "retrieve",
	Short:   "Retrieve the DID document of an existing identifier",
	Example: "algoid retrieve [existing DID]",
	Aliases: []string{"get", "resolve"},
	RunE:    runRetrieveCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "validate",
			Usage:     "validate cryptographic proof received",
			FlagKey:   "resolve.validate",
			ByDefault: false,
			Short:     "v",
		},
	}
	if err := cli.SetupCommandParams(retrieveCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(retrieveCmd)
}

func runRetrieveCmd(_ *cobra.Command, args []string) error {
	// Check params
	if len(args) != 1 {
		return errors.New("you must specify a DID to retrieve")
	}

	// Verify the provided value is a valid DID string
	_, err := did.Parse(args[0])
	if err != nil {
		return err
	}

	// Retrieve record
	log.Info("retrieving record")
	response, err := resolve(args[0])
	if err != nil {
		return fmt.Errorf("failed to resolve DID: %w", err)
	}

	// If no validation is required/supported, print response as-is
	if !viper.GetBool("resolve.validate") {
		log.Warning("skipping validation")
		fmt.Printf("%s\n", response)
		return nil
	}

	// Assume a local record structure
	ir := new(store.IdentifierRecord)
	err = json.Unmarshal(response, ir)
	if err != nil || ir.Document == nil || ir.Proof == nil {
		log.Warning("validation is not supported")
	} else {
		// Restore ID instance
		id, err := did.FromDocument(ir.Document)
		if err != nil {
			return err
		}

		// Verify proof
		key := id.VerificationMethod(ir.Proof.VerificationMethod)
		if key == nil {
			return errors.New("verification method not present in DID document")
		}
		data, err := id.Document(true).NormalizedLD()
		if err != nil {
			return err
		}
		if key.VerifyProof(data, ir.Proof) {
			log.Info("DID document is valid")
		} else {
			log.Warning("DID document is invalid")
		}
	}

	// Print response as output
	fmt.Printf("%s\n", response)
	return nil
}
