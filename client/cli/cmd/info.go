package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var didDetailsCmd = &cobra.Command{
	Use:     "info",
	Short:   "Display the current information available on an existing DID",
	Example: "algoid info [DID reference name]",
	Aliases: []string{"details", "inspect", "view", "doc"},
	RunE:    runDidDetailsCmd,
}

func init() {
	rootCmd.AddCommand(didDetailsCmd)
}

func runDidDetailsCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("you must specify a DID reference name")
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
		return fmt.Errorf("no available record under the provided reference name: %s", name)
	}

	// Present its LD document as output
	info, _ := json.MarshalIndent(id.Document(true), "", "  ")
	fmt.Printf("%s\n", info)
	return nil
}
