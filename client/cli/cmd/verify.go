package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/did"
)

var verifyCmd = &cobra.Command{
	Use:     "verify",
	Short:   "Check the validity of a ProofLD document",
	Example: "algoid verify [proof file] --input \"contents to verify\"",
	RunE:    runVerifyCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "input",
			Usage:     "original contents to run the verification against",
			FlagKey:   "verify.input",
			ByDefault: "",
			Short:     "i",
		},
	}
	if err := cli.SetupCommandParams(verifyCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(verifyCmd)
}

func runVerifyCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("you must provide the signature file to verify")
	}

	// Get input, CLI takes precedence, from standard input otherwise
	input := []byte(viper.GetString("verify.input"))
	if len(input) == 0 {
		input, _ = cli.ReadPipedInput(maxPipeInputSize)
	}
	if len(input) == 0 {
		return errors.New("no input passed in to verify")
	}

	// Load proof file
	log.Info("verifying proof document")
	log.Debug("load signature file")
	entry, err := ioutil.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("failed to read the signature file: %s", err)
	}
	log.Debug("decoding contents")
	proof := &did.ProofLD{}
	if err = json.Unmarshal(entry, proof); err != nil {
		return fmt.Errorf("invalid signature file: %s", err)
	}

	// Validate verification method
	log.Debug("validating proof verification method")
	vm := proof.VerificationMethod
	id, err := did.Parse(vm)
	if err != nil {
		return fmt.Errorf("invalid proof verification method: %s", err)
	}

	// Retrieve subject
	jsDoc, err := resolve(id.String())
	if err != nil {
		return err
	}

	// Decode result obtained from resolve
	doc := new(did.Document)
	result := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(jsDoc), &result); err != nil {
		return fmt.Errorf("invalid DID document received: %s", jsDoc)
	}
	if _, ok := result["document"]; !ok {
		return fmt.Errorf("invalid DID document received: %s", jsDoc)
	}
	if err := json.Unmarshal(result["document"], doc); err != nil {
		return fmt.Errorf("invalid DID document received: %s", jsDoc)
	}

	// Restore peer DID instance
	peer, err := did.FromDocument(doc)
	if err != nil {
		return err
	}

	// Get creator's key
	ck := peer.VerificationMethod(vm)
	if ck == nil {
		return fmt.Errorf("verification method is not available on the DID document: %s", vm)
	}

	// Verify signature
	if !ck.VerifyProof(input, proof) {
		return errors.New("proof is invalid")
	}
	log.Info("proof is valid")
	return nil
}
