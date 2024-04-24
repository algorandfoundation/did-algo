package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:     "activate",
	Short:   "Mark a DID as active",
	Example: "algoid edit activate [DID reference name]",
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("you must specify a DID reference name")
		}

		// Get store handler
		st, err := getClientStore()
		if err != nil {
			return err
		}

		// Get identifier
		name := sanitize.Name(args[0])
		log.Debugf("retrieving entry with reference name: %s", name)
		id, err := st.Get(name)
		if err != nil {
			return fmt.Errorf("no available record under the provided reference name: %s", name)
		}

		// Update metadata
		log.Info("updating metadata")
		md := id.GetMetadata()
		md.Deactivated = false
		md.Updated = time.Now().UTC().Format(time.RFC3339)
		if err := id.AddMetadata(md); err != nil {
			return err
		}

		// Update record
		log.Info("updating local record")
		return st.Update(name, id)
	},
}

func init() {
	editCmd.AddCommand(activateCmd)
}
