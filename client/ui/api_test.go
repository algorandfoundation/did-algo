package ui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/algorand/go-algorand-sdk/client/kmd"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/crypto"

	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorandfoundation/did-algo/client/internal"
	"github.com/algorandfoundation/did-algo/client/store"

	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/log"
)

var algodClient *algod.Client
var logger = log.WithZero(log.ZeroOptions{PrettyPrint: true})
var localStore *store.LocalStore
var profile internal.NetworkProfile

func getKmdClient() (kmd.Client, error) {
	kmdClient, err := kmd.MakeClient(
		"http://localhost:4002",
		strings.Repeat("a", 64),
	)

	return kmdClient, err
}

func getSandboxAccounts() ([]crypto.Account, error) {
	client, err := getKmdClient()
	if err != nil {
		return nil, fmt.Errorf("Failed to create kmd client: %+v", err)
	}

	resp, err := client.ListWallets()
	if err != nil {
		return nil, fmt.Errorf("Failed to list wallets: %+v", err)
	}

	var walletId string
	for _, wallet := range resp.Wallets {
		if wallet.Name == "unencrypted-default-wallet" {
			walletId = wallet.ID
		}
	}

	if walletId == "" {
		return nil, fmt.Errorf("No wallet named %s", "unencrypted-default-wallet")
	}

	whResp, err := client.InitWalletHandle(walletId, "")
	if err != nil {
		return nil, fmt.Errorf("Failed to init wallet handle: %+v", err)
	}

	addrResp, err := client.ListKeys(whResp.WalletHandleToken)
	if err != nil {
		return nil, fmt.Errorf("Failed to list keys: %+v", err)
	}

	var accts []crypto.Account
	for _, addr := range addrResp.Addresses {
		expResp, err := client.ExportKey(whResp.WalletHandleToken, "", addr)
		if err != nil {
			return nil, fmt.Errorf("Failed to export key: %+v", err)
		}

		acct, err := crypto.AccountFromPrivateKey(expResp.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("Failed to create account from private key: %+v", err)
		}

		accts = append(accts, acct)
	}

	return accts, nil
}

func req(t *testing.T, method string, endpoint string, reqBody string) *http.Response {
	jsonBody := []byte(reqBody)
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(method, fmt.Sprintf("http://localhost:9090/%s", endpoint), bodyReader)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)

	return res
}

func TestMain(m *testing.M) {
	home, err := filepath.Abs(fmt.Sprintf("./.test/%d", time.Now().Unix()))
	if err != nil {
		panic(err)
	}

	localStore, err = store.NewLocalStore(home)
	if err != nil {
		panic(err)
	}

	algodClient, err = algod.MakeClient("http://localhost:4001", strings.Repeat("a", 64))
	if err != nil {
		panic(err)
	}

	status, err := algodClient.Status().Do(context.Background())

	// type NetworkProfile struct {
	// 	// Profile name.
	// 	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// 	// Algod node address.
	// 	Node string `json:"node" yaml:"node" mapstructure:"node"`

	// 	// Algod node access token.
	// 	NodeToken string `json:"node_token,omitempty" yaml:"node_token,omitempty" mapstructure:"node_token"`

	// 	// Application ID for the AlgoDID storage smart contract.
	// 	AppID uint `json:"app_id" yaml:"app_id" mapstructure:"app_id"`

	// 	// Storage contract provider server, if any.
	// 	StoreProvider string `json:"store_provider,omitempty" yaml:"store_provider,omitempty" mapstructure:"store_provider"`
	// }

	block, err := algodClient.Block(status.LastRound).Do(context.Background())
	if err != nil {
		panic(err)
	}

	profile = internal.NetworkProfile{
		Name:      "custom",
		Node:      "http://localhost:4001",
		NodeToken: strings.Repeat("a", 64),
		AppID:     uint(block.TxnCounter) + 2,
	}

	fmt.Printf("Using app ID: %d\n", profile.AppID)

	profiles := []*internal.NetworkProfile{&profile}

	// Get network client
	cl, err := internal.NewAlgoClient(profiles, logger)
	if err != nil {
		panic(err)
	}

	srv, err := LocalAPIServer(localStore, cl, logger)

	logger.Debug("starting local API server on localhost:9090")
	go func() {
		_ = srv.Start()
	}()

	m.Run()
}

func TestReady(t *testing.T) {
	res := req(t, http.MethodGet, "ready", "")

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestList(t *testing.T) {
	res := req(t, http.MethodGet, "list", "")

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, "[]\n", string(body))
}

func TestRegister(t *testing.T) {
	res := req(t, http.MethodPost, "register", `{"name": "TestRegister", "recovery_key": "test", "network": "custom"}`)

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestListAfterRegister(t *testing.T) {
	res := req(t, http.MethodPost, "register", `{"name": "TestListAfterRegister", "recovery_key": "test", "network": "custom"}`)

	require.Equal(t, http.StatusOK, res.StatusCode)

	res = req(t, http.MethodGet, "list", "")

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Regexp(t, `{"name":"TestListAfterRegister","did":"did:algo:custom:app:`, string(body))
}

func TestUpdate(t *testing.T) {
	res := req(t, http.MethodPost, "register", `{"name": "TestUpdate", "recovery_key": "test", "network": "custom"}`)

	require.Equal(t, http.StatusOK, res.StatusCode)

	// Get account

	mn, err := localStore.OpenWallet("TestUpdate", "test")
	require.NoError(t, err)

	sk, err := mnemonic.ToPrivateKey(mn)
	require.NoError(t, err)

	acct, err := crypto.AccountFromPrivateKey(sk)
	require.NoError(t, err)

	signer := transaction.BasicAccountTransactionSigner{Account: acct}

	require.NoError(t, err)

	// Fund account

	localAccounts, err := getSandboxAccounts()
	require.NoError(t, err)

	funder := localAccounts[0]

	sp, err := algodClient.SuggestedParams().Do(context.Background())
	require.NoError(t, err)

	fundTxn, err := transaction.MakePaymentTxn(funder.Address.String(), acct.Address.String(), 1_000_000, nil, "", sp)
	require.NoError(t, err)

	txid, stxn, err := crypto.SignTransaction(funder.PrivateKey, fundTxn)
	require.NoError(t, err)

	_, err = algodClient.SendRawTransaction(stxn).Do(context.Background())
	require.NoError(t, err)

	_, err = transaction.WaitForConfirmation(algodClient, txid, 5, context.Background())
	require.NoError(t, err)

	// Create the app

	createdAppId, err := internal.CreateApp(algodClient, internal.LoadContract(), acct.Address, signer)
	require.NoError(t, err)
	require.Equal(t, uint64(profile.AppID), createdAppId)
	fmt.Printf("Created app ID: %d\n", createdAppId)

	appInfo, err := algodClient.GetApplicationByID(createdAppId).Do(context.Background())
	require.NoError(t, err)
	require.False(t, appInfo.Deleted)

	// Sync with the network

	res = req(t, http.MethodPost, "update", `{"name": "TestUpdate", "passphrase": "test"}`)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	did, err := localStore.Get("TestUpdate")
	require.NoError(t, err)

	resolvedDid, err := internal.ResolveDID(createdAppId, acct.PublicKey, algodClient, "custom")
	require.NoError(t, err)

	fmt.Println(string(resolvedDid))

	require.Regexp(t, fmt.Sprintf(`"id":"%s"`, did.DID()), string(resolvedDid))
}
