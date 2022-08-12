package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/did"
)

var addKeyCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a new cryptographic key for the DID",
	Example: "algoid edit key add [DID reference name] --name my-new-key --type ed --authentication",
	RunE:    runAddKeyCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "name",
			Usage:     "name to be assigned to the newly added key",
			FlagKey:   "key-add.name",
			ByDefault: "key-#",
			Short:     "n",
		},
		{
			Name:      "type",
			Usage:     "type of cryptographic key: RSA (rsa), Ed25519 (ed) or secp256k1 (koblitz)",
			FlagKey:   "key-add.type",
			ByDefault: "ed",
			Short:     "t",
		},
		{
			Name:      "authentication",
			Usage:     "enable this key for authentication purposes",
			FlagKey:   "key-add.authentication",
			ByDefault: false,
			Short:     "a",
		},
	}
	if err := cli.SetupCommandParams(addKeyCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	keyCmd.AddCommand(addKeyCmd)
}

// nolint: gocyclo
func runAddKeyCmd(_ *cobra.Command, args []string) error {
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
	log.Info("adding new key")
	log.Debugf("retrieving entry with reference name: %s", name)
	id, err := st.Get(name)
	if err != nil {
		return fmt.Errorf("no available record under the provided reference name: %s", name)
	}

	// Sanitize key name
	log.Debug("validating parameters")
	keyName := viper.GetString("key-add.name")
	if strings.Count(keyName, "#") > 1 {
		return errors.New("invalid key name")
	}
	if strings.Count(keyName, "#") == 1 {
		keyName = strings.Replace(keyName, "#", fmt.Sprintf("%d", len(id.VerificationMethods())+1), 1)
	}
	keyName = sanitize.Name(keyName)

	// Set key type
	var keyType did.KeyType
	switch viper.GetString("key-add.type") {
	case "ed":
		keyType = did.KeyTypeEd
	case "rsa":
		keyType = did.KeyTypeRSA
	case "koblitz":
		keyType = did.KeyTypeSecp256k1
	default:
		return errors.New("invalid key type")
	}

	// Add key
	log.Debugf("adding new key with name: %s", keyName)
	if err = id.AddNewVerificationMethod(keyName, keyType); err != nil {
		return fmt.Errorf("failed to add new key: %w", err)
	}
	if viper.GetBool("key-add.authentication") {
		log.Info("setting new key as authentication mechanism")
		if err = id.AddVerificationRelationship(id.GetReference(keyName), did.AuthenticationVM); err != nil {
			return fmt.Errorf("failed to establish key for authentication purposes: %w", err)
		}
	}

	// Update record
	log.Info("updating local record")
	return st.Update(name, id)
}
