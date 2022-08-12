package cmd

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/crypto/shamir"
	"go.bryk.io/pkg/did"
	xlog "go.bryk.io/pkg/log"
)

var registerCmd = &cobra.Command{
	Use:     "register",
	Short:   "Creates a new DID locally",
	Example: "algoid register [DID reference name]",
	Aliases: []string{"create", "new"},
	RunE:    runRegisterCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "passphrase",
			Usage:     "set a passphrase as recovery method for the primary key",
			FlagKey:   "register.passphrase",
			ByDefault: false,
			Short:     "p",
		},
		{
			Name:      "secret-sharing",
			Usage:     "number of shares and threshold value: shares,threshold",
			FlagKey:   "register.secret-sharing",
			ByDefault: "3,2",
			Short:     "s",
		},
		{
			Name:      "tag",
			Usage:     "tag value for the identifier instance",
			FlagKey:   "register.tag",
			ByDefault: "",
			Short:     "t",
		},
		{
			Name:      "method",
			Usage:     "method value for the identifier instance",
			FlagKey:   "register.method",
			ByDefault: "algo",
			Short:     "m",
		},
	}
	if err := cli.SetupCommandParams(registerCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(registerCmd)
}

func runRegisterCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("a reference name for the DID is required")
	}
	name := sanitize.Name(args[0])

	// Get store handler
	st, err := getClientStore()
	if err != nil {
		return err
	}

	// Check for duplicates
	dup, _ := st.Get(name)
	if dup != nil {
		return fmt.Errorf("there's already a DID with reference name: %s", name)
	}

	// Get key secret from the user
	log.Info("obtaining secret material for the master private key")
	secret, err := getSecret(name)
	if err != nil {
		return err
	}

	// Generate master key from available secret
	masterKey, err := keyFromMaterial(secret)
	if err != nil {
		return err
	}
	defer masterKey.Destroy()
	pk := make([]byte, 64)
	copy(pk, masterKey.PrivateKey())

	// Generate base identifier instance
	method := viper.GetString("register.method")
	tag := viper.GetString("register.tag")
	log.WithFields(xlog.Fields{
		"method": method,
		"tag":    tag,
	}).Info("generating new identifier")
	id, err := did.NewIdentifierWithMode(method, tag, did.ModeUUID)
	if err != nil {
		return err
	}
	log.Debug("adding master key")
	if err = id.AddVerificationMethod("master", pk, did.KeyTypeEd); err != nil {
		return err
	}
	log.Debug("setting master key as authentication mechanism")
	if err = id.AddVerificationRelationship(id.GetReference("master"), did.AuthenticationVM); err != nil {
		return err
	}

	// Save instance in the store
	log.Info("adding entry to local store")
	return st.Save(name, id)
}

func getSecret(name string) ([]byte, error) {
	// User provided passphrase
	if viper.GetBool("register.passphrase") {
		secret, err := readSecretValue("Enter a secure passphrase")
		if err != nil {
			return nil, err
		}
		confirmation, err := readSecretValue("Confirm the provided value")
		if err != nil {
			return nil, err
		}
		if secret != confirmation {
			return nil, errors.New("the values provided are not equal")
		}
		return []byte(secret), nil
	}

	// Shared secret
	secret := make([]byte, 128)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	// Spilt secret and save shares to local files
	shares, err := splitSecret(secret, viper.GetString("register.secret-sharing"))
	if err != nil {
		return nil, err
	}
	for i, k := range shares {
		share := fmt.Sprintf("%s.share_%d.bin", name, i+1)
		if err := os.WriteFile(share, k, 0400); err != nil {
			return nil, fmt.Errorf("failed to save share '%s': %w", share, err)
		}
	}
	return secret, nil
}

func splitSecret(secret []byte, conf string) ([][]byte, error) {
	// Load configuration
	sssConf := strings.Split(conf, ",")
	if len(sssConf) != 2 {
		return nil, errors.New("invalid secret sharing configuration value")
	}

	// Validate configuration
	shares, err := strconv.Atoi(sssConf[0])
	if err != nil {
		return nil, fmt.Errorf("invalid number shares: %s", sssConf[0])
	}
	threshold, err := strconv.Atoi(sssConf[1])
	if err != nil {
		return nil, fmt.Errorf("invalid threshold value: %s", sssConf[1])
	}
	if threshold >= shares {
		return nil, fmt.Errorf("threshold '(%d)' should be smaller than shares '(%d)'", threshold, shares)
	}

	// Split secret
	return shamir.Split(secret, shares, threshold)
}
