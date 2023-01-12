package cmd

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/algorandfoundation/did-algo/client/internal"
	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/did"
	xlog "go.bryk.io/pkg/log"
)

var syncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Publish a DID instance to the processing network",
	Example: "algoid sync [DID reference name]",
	Aliases: []string{"publish", "update", "upload", "push"},
	RunE:    runSyncCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "key",
			Usage:     "cryptographic key to use for the sync operation",
			FlagKey:   "sync.key",
			ByDefault: "master",
			Short:     "k",
		},
		{
			Name:      "pow",
			Usage:     "set the required request ticket difficulty level",
			FlagKey:   "client.pow",
			ByDefault: 24,
			Short:     "p",
		},
	}
	if err := cli.SetupCommandParams(syncCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(syncCmd)
}

func runSyncCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("you must specify a DID reference name")
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

	// Get selected key for the sync operation
	key, err := getSyncKey(id)
	if err != nil {
		return err
	}
	log.Debugf("key selected for the operation: %s", key.ID)

	// Generate request ticket
	log.Infof("publishing: %s", name)
	ticket, err := getRequestTicket(id, key)
	if err != nil {
		return err
	}

	// Get client connection
	conf := new(internal.ClientSettings)
	if err := viper.UnmarshalKey("client", conf); err != nil {
		return err
	}
	if err := conf.Validate(); err != nil {
		return err
	}
	conn, err := getClientConnection(conf)
	if err != nil {
		return fmt.Errorf("failed to establish connection: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Build request
	req := &protoV1.ProcessRequest{
		Ticket: ticket,
	}

	// Submit request
	log.Info("submitting request to the network")
	client := protoV1.NewAgentAPIClient(conn)
	res, err := client.Process(context.Background(), req)
	if err != nil {
		return fmt.Errorf("network return an error: %w", err)
	}
	log.Debugf("request status: %v", res.Ok)
	if res.Identifier != "" {
		log.Info("identifier: ", res.Identifier)
	}
	if !res.Ok {
		return nil
	}

	// Update local record if sync was successful
	return st.Update(name, id)
}

func getRequestTicket(id *did.Identifier, key *did.VerificationKey) (*protoV1.Ticket, error) {
	diff := uint(viper.GetInt("client.pow"))
	log.WithFields(xlog.Fields{"pow": diff}).Info("generating request ticket")

	// Create new ticket
	ticket, err := protoV1.NewTicket(id, key.ID)
	if err != nil {
		return nil, err
	}

	// Solve PoW challenge
	start := time.Now()
	challenge := ticket.Solve(context.Background(), diff)
	log.Debugf("ticket obtained: %s", challenge)
	log.Debugf("time: %s (rounds completed %d)", time.Since(start), ticket.Nonce())
	ch, _ := hex.DecodeString(challenge)

	// Sign ticket
	if ticket.Signature, err = key.Sign(ch); err != nil {
		return nil, fmt.Errorf("failed to generate request ticket: %w", err)
	}

	// Verify on client's side
	if err = ticket.Verify(diff); err != nil {
		return nil, fmt.Errorf("failed to verify ticket: %w", err)
	}

	return ticket, nil
}

func getSyncKey(id *did.Identifier) (*did.VerificationKey, error) {
	// Get selected key for the sync operation
	key := id.VerificationMethod(viper.GetString("sync.key"))
	if key == nil {
		return nil, errors.New("invalid key selected")
	}

	// Verify the key is enabled for authentication
	isAuth := false
	for _, k := range id.GetVerificationRelationship(did.AuthenticationVM) {
		if k == key.ID {
			isAuth = true
			break
		}
	}
	if !isAuth {
		return nil, errors.New("the key selected is not enabled for authentication purposes")
	}
	return key, nil
}
