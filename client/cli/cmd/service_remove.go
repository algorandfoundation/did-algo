package cmd

import (
	"errors"
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var removeServiceCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove an existing service entry for the DID",
	Example: "algoid edit service remove [DID reference name] [service name]",
	Aliases: []string{"rm"},
	RunE:    runRemoveServiceCmd,
}

func init() {
	serviceCmd.AddCommand(removeServiceCmd)
}

func runRemoveServiceCmd(_ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("you must specify [DID reference name] [service name]")
	}

	// Get store handler
	st, err := getClientStore()
	if err != nil {
		return err
	}

	// Get identifier
	name := sanitize.Name(args[0])
	log.Info("removing existing service")
	log.Debugf("retrieving entry with reference name: %s", name)
	id, err := st.Get(name)
	if err != nil {
		return fmt.Errorf("no available record under the provided reference name: %s", name)
	}

	// Remove service
	sName := sanitize.Name(args[1])
	log.Debugf("deleting service with name: %s", sName)
	if err = id.RemoveService(sName); err != nil {
		return fmt.Errorf("failed to remove service: %s", sName)
	}

	// Update record
	log.Info("updating local record")
	return st.Update(name, id)
}
