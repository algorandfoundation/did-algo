package cmd

import (
	"errors"
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"go.bryk.io/pkg/did"
)

var removeKeyCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove an existing cryptographic key for the DID",
	Example: "algoid edit key remove [DID reference name] [key name]",
	Aliases: []string{"rm"},
	RunE:    runRemoveKeyCmd,
}

func init() {
	keyCmd.AddCommand(removeKeyCmd)
}

func runRemoveKeyCmd(_ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("you must specify [DID reference name] [key name]")
	}

	// Get store handler
	st, err := getClientStore()
	if err != nil {
		return err
	}

	// Get identifier
	name := sanitize.Name(args[0])
	keyName := sanitize.Name(args[1])
	log.Info("removing existing key")
	log.Debugf("retrieving entry with reference name: %s", name)
	id, err := st.Get(name)
	if err != nil {
		return fmt.Errorf("no available record under the provided reference name: %s", name)
	}

	// Remove key
	log.Debug("validating parameters")
	if len(id.VerificationMethods()) >= 2 {
		_ = id.RemoveVerificationRelationship(id.GetReference(keyName), did.AuthenticationVM)
	}
	if err = id.RemoveVerificationMethod(keyName); err != nil {
		return fmt.Errorf("failed to remove key: %s", keyName)
	}

	// Update record
	log.Debug("key removed")
	log.Info("updating local record")
	return st.Update(name, id)
}
