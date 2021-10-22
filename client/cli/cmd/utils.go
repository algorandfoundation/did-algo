package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/algorandfoundation/did-algo/client/store"
	"github.com/algorandfoundation/did-algo/info"
	"github.com/algorandfoundation/did-algo/resolver"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/crypto/ed25519"
	"go.bryk.io/pkg/net/rpc"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc"
)

// When reading contents from standard input a maximum of 4MB is expected
const maxPipeInputSize = 4096

// Securely expand the provided secret material
func expand(secret []byte, size int, info []byte) ([]byte, error) {
	salt := make([]byte, sha256.Size)
	buf := make([]byte, size)
	h := hkdf.New(sha3.New256, secret, salt[:], info)
	if _, err := io.ReadFull(h, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// Restore key pair from the provided material
func keyFromMaterial(material []byte) (*ed25519.KeyPair, error) {
	m, err := expand(material, 32, nil)
	if err != nil {
		return nil, err
	}
	seed := [32]byte{}
	copy(seed[:], m)
	return ed25519.FromSeed(seed[:])
}

// Accessor to the local storage handler
func getClientStore() (*store.LocalStore, error) {
	return store.NewLocalStore(viper.GetString("client.home"))
}

// Get an RPC network connection
func getClientConnection() (*grpc.ClientConn, error) {
	node := viper.GetString("client.node")
	log.Infof("establishing connection to network agent: %s", node)
	timeout := viper.GetInt("client.timeout")
	opts := []rpc.ClientOption{
		rpc.WaitForReady(),
		rpc.WithUserAgent(fmt.Sprintf("algoid-client/%s", info.CoreVersion)),
		rpc.WithTimeout(time.Duration(timeout) * time.Second),
	}
	if viper.GetBool("client.tls") {
		opts = append(opts, rpc.WithClientTLS(rpc.ClientTLSConfig{IncludeSystemCAs: true}))
	}
	if override := viper.GetString("client.override"); override != "" {
		opts = append(opts, rpc.WithServerNameOverride(override))
	}
	return rpc.NewClientConnection(node, opts...)
}

// Use the global resolver to obtain the DID document for the requested
// identifier.
func resolve(id string) ([]byte, error) {
	var conf []*resolver.Provider
	if err := viper.UnmarshalKey("resolver", &conf); err != nil {
		return nil, fmt.Errorf("invalid resolver configuration: %s", err)
	}
	return resolver.Get(id, conf)
}

// Get algo node client
func algodClient() (*algod.Client, error) {
	address := viper.GetString("agent.network.algod.address")
	token := viper.GetString("agent.network.algod.token")
	log.WithField("address", address).Info("connecting to algod node")
	return algod.MakeClient(address, token)
}

// Get algo indexer client
func indexerClient() (*indexer.Client, error) {
	address := viper.GetString("agent.network.indexer.address")
	token := viper.GetString("agent.network.indexer.token")
	log.WithField("address", address).Info("connecting to indexer node")
	return indexer.MakeClient(address, token)
}
