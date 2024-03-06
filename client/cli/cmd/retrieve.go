package cmd

import (
	"encoding/json"
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

	// Get network client
	cl, err := getAlgoClient()
	if err != nil {
		return err
	}

	// Retrieve record
	log.Info("retrieving record")
	doc, err := cl.Resolve(args[0])
	if err != nil {
		return err
	}

	// Pretty-print retrieved document
	log.Warning("skipping validation")
	output, _ := json.MarshalIndent(doc, "", "  ")
	fmt.Printf("%s\n", output)
	return nil
}
