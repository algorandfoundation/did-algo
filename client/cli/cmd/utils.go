package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/algorandfoundation/did-algo/client/store"
	"github.com/algorandfoundation/did-algo/info"
	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/crypto/ed25519"
	"go.bryk.io/pkg/did"
	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/net/rpc"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc"
)

// When reading contents from standard input a maximum of 4MB is expected.
const maxPipeInputSize = 4096

// Securely expand the provided secret material.
func expand(secret []byte, size int, info []byte) ([]byte, error) {
	salt := make([]byte, sha256.Size)
	buf := make([]byte, size)
	h := hkdf.New(sha3.New256, secret, salt[:], info)
	if _, err := io.ReadFull(h, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// Restore key pair from the provided material.
func keyFromMaterial(material []byte) (*ed25519.KeyPair, error) {
	m, err := expand(material, 32, nil)
	if err != nil {
		return nil, err
	}
	seed := [32]byte{}
	copy(seed[:], m)
	return ed25519.FromSeed(seed[:])
}

// Accessor to the local storage handler.
func getClientStore() (*store.LocalStore, error) {
	return store.NewLocalStore(viper.GetString("client.home"))
}

// Get an RPC network connection.
func getClientConnection(conf *internal.ClientSettings) (*grpc.ClientConn, error) {
	log.Infof("establishing connection to network agent: %s", conf.Node)
	opts := []rpc.ClientOption{
		rpc.WaitForReady(),
		rpc.WithUserAgent(fmt.Sprintf("algoid-client/%s", info.CoreVersion)),
		rpc.WithTimeout(time.Duration(conf.Timeout) * time.Second),
	}
	if conf.Insecure {
		log.Warning("using an insecure connection")
	} else {
		opts = append(opts, rpc.WithClientTLS(rpc.ClientTLSConfig{IncludeSystemCAs: true}))
	}
	if conf.Override != "" {
		log.WithField("override", conf.Override).Warning("using server name override")
		opts = append(opts, rpc.WithServerNameOverride(conf.Override))
	}
	return rpc.NewClientConnection(conf.Node, opts...)
}

// Get algo node client.
func algodClient() (*algod.Client, error) {
	address := viper.GetString("agent.network.algod.address")
	token := viper.GetString("agent.network.algod.token")
	log.WithField("address", address).Info("connecting to algod node")
	return algod.MakeClient(address, token)
}

// Get algo indexer client.
func indexerClient() (*indexer.Client, error) {
	address := viper.GetString("agent.network.indexer.address")
	token := viper.GetString("agent.network.indexer.token")
	log.WithField("address", address).Info("connecting to indexer node")
	return indexer.MakeClient(address, token)
}

// Use internal RPC client to obtain the DID document for the requested
// identifier.
func resolve(id string) ([]byte, error) {
	// NOTE: future versions could use the DIF universal resolver
	// https://dev.uniresolver.io/1.0/identifiers/{id}

	// Get client connection
	conf := new(internal.ClientSettings)
	if err := viper.UnmarshalKey("client", conf); err != nil {
		return nil, err
	}
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	conn, err := getClientConnection(conf)
	if err != nil {
		return nil, errors.Errorf("failed to establish connection: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Submit query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ID, _ := did.Parse(id)
	cl := protoV1.NewAgentAPIClient(conn)
	res, err := cl.Query(ctx, &protoV1.QueryRequest{
		Method:  ID.Method(),
		Subject: ID.Subject(),
	})
	if err != nil {
		return nil, err
	}

	// Return DID document
	return res.Document, nil
}
