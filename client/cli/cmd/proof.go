package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
)

var proofCmd = &cobra.Command{
	Use:     "proof",
	Short:   "Produce a linked digital proof document",
	Example: "algoid proof [DID reference name] --input \"contents to sign\"",
	Aliases: []string{"sign"},
	RunE:    runProofCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "input",
			Usage:     "contents to sign",
			FlagKey:   "proof.input",
			ByDefault: "",
			Short:     "i",
		},
		{
			Name:      "key",
			Usage:     "key to use to produce the proof",
			FlagKey:   "proof.key",
			ByDefault: "master",
			Short:     "k",
		},
		{
			Name:      "domain",
			Usage:     "domain value to use",
			FlagKey:   "proof.domain",
			ByDefault: didDomainValue,
			Short:     "d",
		},
		{
			Name:      "purpose",
			Usage:     "specific intent for the proof",
			FlagKey:   "proof.purpose",
			ByDefault: "authentication",
			Short:     "p",
		},
	}
	if err := cli.SetupCommandParams(proofCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(proofCmd)
}

func runProofCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("you must specify a DID reference name")
	}

	// Get input, CLI takes precedence, from standard input otherwise
	input := []byte(viper.GetString("proof.input"))
	if len(input) == 0 {
		input, _ = cli.ReadPipedInput(maxPipeInputSize)
	}
	if len(input) == 0 {
		return errors.New("no input passed in to sign")
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

	// Get key
	key := id.VerificationMethod(viper.GetString("proof.key"))
	if key == nil {
		return fmt.Errorf("selected key is not available on the DID: %s", viper.GetString("proof.key"))
	}

	// Produce proof
	purpose := viper.GetString("proof.purpose")
	domain := viper.GetString("proof.domain")
	pld, err := key.ProduceProof(input, purpose, domain)
	if err != nil {
		return fmt.Errorf("failed to produce proof: %s", err)
	}
	js, err := json.MarshalIndent(pld, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to produce proof: %s", err)
	}
	fmt.Printf("%s\n", js)
	return nil
}
