package cmd

import (
	"errors"
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/did"
)

var walletConnectCmd = &cobra.Command{
	Use:     "connect",
	Aliases: []string{"link", "ln"},
	RunE:    runWalletConnectCmd,
	Example: "algoid wallet connect [wallet-name] [did-name]",
	Short:   "Connect your ALGO wallet to a DID",
	Long: `Connect your ALGO wallet to a DID

Connecting your ALGO wallet to a DID will allow other users to
discover your ALGO address when resolving your identifier.

Effectively connecting your ID to a highly secure and efficient
payments channel. Additionally, your counterparties might also
discover or request your credentials when/if required to perform
certain transactions.`,
}

func init() {
	params := []cli.Param{
		{
			Name:      "network",
			Usage:     "Algorand network to use",
			FlagKey:   "wallet-connect.network",
			ByDefault: "testnet",
			Short:     "n",
		},
		{
			Name:      "asset",
			Usage:     "Asset advertised for the connection",
			FlagKey:   "wallet-connect.asset",
			ByDefault: "ALGO",
			Short:     "a",
		},
	}
	if err := cli.SetupCommandParams(walletConnectCmd, params); err != nil {
		panic(err)
	}
	walletCmd.AddCommand(walletConnectCmd)
}

func runWalletConnectCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("missing required parameters")
	}

	// Get parameters
	name := sanitize.Name(args[0])
	didName := sanitize.Name(args[1])
	wp, err := readSecretValue("enter wallet's passphrase")
	if err != nil {
		return err
	}

	// Get wallet address
	address, err := getWalletAddress(name, wp)
	if err != nil {
		return err
	}

	// Get identifier instance
	store, err := getClientStore()
	if err != nil {
		return err
	}
	id, err := store.Get(didName)
	if err != nil {
		return fmt.Errorf("no available record under the provided reference name: %s", name)
	}

	// Get service entry
	svc := id.Service("algo-connect")
	if svc == nil {
		svc = getServiceEntry()
	}

	// Retrieve existing address list and add new one
	addresses := []algoDestination{}
	ext := did.Extension{
		ID:      "algo-address",
		Version: "0.1.0",
	}
	if err := svc.GetExtension(ext.ID, ext.Version, &addresses); err != nil {
		return err
	}
	for _, entry := range addresses {
		if entry.Address == address {
			return fmt.Errorf("address already linked to DID: %s", address)
		}
	}
	addresses = append(addresses, algoDestination{
		Address: address,
		Network: viper.GetString("wallet-connect.network"),
		Asset:   viper.GetString("wallet-connect.asset"),
	})
	ext.Data = addresses
	svc.AddExtension(ext)

	// Update service entry
	_ = id.RemoveService("algo-connect")
	if err := id.AddService(svc); err != nil {
		return err
	}

	// Register custom context
	id.RegisterContext("https://did-ns.aidtech.network/v1")

	// Update record
	log.Info("updating local DID record")
	return store.Update(didName, id)
}

func getServiceEntry() *did.ServiceEndpoint {
	return &did.ServiceEndpoint{
		ID:       "algo-connect",
		Type:     "did.algorand.foundation.ExternalService",
		Endpoint: "https://did.algorand.foundation",
		Extensions: []did.Extension{
			{
				ID:      "algo-address",
				Version: "0.1.0",
				Data:    []algoDestination{},
			},
		},
	}
}
