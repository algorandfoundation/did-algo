package cmd

import (
	"errors"
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"go.bryk.io/pkg/did"
)

var walletDisconnectCmd = &cobra.Command{
	Use:     "disconnect",
	Aliases: []string{"unlink"},
	Short:   "Remove a linked ALGO address from your DID",
	Example: "algoid wallet disconnect [did-name] [algo-address]",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("missing required parameters")
		}

		// Get parameters
		didName := sanitize.Name(args[0])
		addr := args[1]

		// Get identifier instance
		store, err := getClientStore()
		if err != nil {
			return err
		}
		id, err := store.Get(didName)
		if err != nil {
			return fmt.Errorf("no available record under the provided reference name: %s", didName)
		}

		// Get service entry
		svc := id.Service("algo-connect")
		if svc == nil {
			return fmt.Errorf("no ALGO address linked to the DID")
		}

		// Get list of linked address
		oldList := []algoDestination{}
		ext := did.Extension{
			ID:      "algo-address",
			Version: "0.1.0",
		}
		if err := svc.GetExtension(ext.ID, ext.Version, &oldList); err != nil {
			return err
		}

		// Filter out provided address
		newList := []algoDestination{}
		for _, entry := range oldList {
			if entry.Address != addr {
				newList = append(newList, entry)
			}
		}
		ext.Data = newList

		// Update service entry
		for i, e := range svc.Extensions {
			if e.ID == ext.ID {
				svc.Extensions = append(svc.Extensions[:i], svc.Extensions[i+1:]...)
				break
			}
		}
		svc.AddExtension(ext)
		_ = id.RemoveService("algo-connect")
		if err = id.AddService(svc); err != nil {
			return err
		}

		// Update record
		log.Info("updating local DID record")
		if err = store.Update(didName, id); err != nil {
			return err
		}
		log.Info("address was successfully unlinked: ", addr)
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletDisconnectCmd)
}
