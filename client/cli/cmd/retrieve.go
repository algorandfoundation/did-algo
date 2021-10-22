package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
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
		return fmt.Errorf("failed to resolve DID: %s", err)
	}
	fmt.Printf("%s\n", response)
	return nil
}
