package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/did"
)

var addServiceCmd = &cobra.Command{
	Use:     "add",
	Short:   "Register a new service entry for the DID",
	Example: "algoid edit service add [DID] --name my-service --endpoint https://www.agency.com/user_id",
	RunE:    runAddServiceCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "name",
			Usage:     "service's reference name",
			FlagKey:   "service-add.name",
			ByDefault: "external-service-#",
			Short:     "n",
		},
		{
			Name:      "type",
			Usage:     "type identifier for the service handler",
			FlagKey:   "service-add.type",
			ByDefault: "did.algorand.foundation.ExternalService",
			Short:     "t",
		},
		{
			Name:      "endpoint",
			Usage:     "main URL to access the service",
			FlagKey:   "service-add.endpoint",
			ByDefault: "",
			Short:     "e",
		},
	}
	if err := cli.SetupCommandParams(addServiceCmd, params); err != nil {
		panic(err)
	}
	serviceCmd.AddCommand(addServiceCmd)
}

func runAddServiceCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("you must specify a DID reference name")
	}
	if strings.TrimSpace(viper.GetString("service-add.endpoint")) == "" {
		return errors.New("service endpoint is required")
	}

	// Get store handler
	st, err := getClientStore()
	if err != nil {
		return err
	}

	// Get identifier
	name := sanitize.Name(args[0])
	log.Info("adding new service")
	log.Debugf("retrieving entry with reference name: %s", name)
	id, err := st.Get(name)
	if err != nil {
		return fmt.Errorf("no available record under the provided reference name: %s", name)
	}

	// Validate service data
	log.Debug("validating parameters")
	svc := &did.ServiceEndpoint{
		ID:       viper.GetString("service-add.name"),
		Type:     viper.GetString("service-add.type"),
		Endpoint: viper.GetString("service-add.endpoint"),
	}
	if strings.Count(svc.ID, "#") > 1 {
		return errors.New("invalid service name")
	}
	if strings.Count(svc.ID, "#") == 1 {
		svc.ID = strings.Replace(svc.ID, "#", fmt.Sprintf("%d", len(id.Services())+1), 1)
	}
	svc.ID = sanitize.Name(svc.ID)
	if _, err = url.ParseRequestURI(svc.Endpoint); err != nil {
		return fmt.Errorf("invalid service endpoint: %s", svc.Endpoint)
	}

	// Add service
	log.Debugf("registering service with id: %s", svc.ID)
	if err = id.AddService(svc); err != nil {
		return fmt.Errorf("failed to add new service: %s", err)
	}

	// Update record
	log.Info("updating local record")
	return st.Update(name, id)
}
