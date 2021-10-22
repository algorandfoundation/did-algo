package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	ac "github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	at "github.com/algorand/go-algorand-sdk/types"
	protov1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"google.golang.org/protobuf/types/known/emptypb"
)

var walletPayCmd = &cobra.Command{
	Use:     "pay",
	Short:   "Create and submit a new transaction",
	Aliases: []string{"txn", "send"},
	Example: "algoid wallet pay [wallet-name]",
	RunE:    runWalletPayCmd,
}

func init() {
	params := []cli.Param{
		{
			Name:      "to",
			Usage:     "Receiver address",
			FlagKey:   "tx.to",
			ByDefault: "",
			Short:     "r",
		},
		{
			Name:      "amount",
			Usage:     "Transaction amount",
			FlagKey:   "tx.amount",
			ByDefault: 0,
			Short:     "a",
		},
		{
			Name:      "submit",
			Usage:     "Submit transaction to the network",
			FlagKey:   "tx.submit",
			ByDefault: false,
			Short:     "s",
		},
	}
	if err := cli.SetupCommandParams(walletPayCmd, params); err != nil {
		panic(err)
	}
	walletCmd.AddCommand(walletPayCmd)
}

func runWalletPayCmd(cmd *cobra.Command, args []string) (err error) {
	// Get parameters
	wallet, receiver, amount, err := getTxParameters(args)
	if err != nil {
		return err
	}
	wp, err := readSecretValue("enter wallet's passphrase")
	if err != nil {
		return err
	}

	// Get local store handler
	store, err := getClientStore()
	if err != nil {
		return err
	}

	// Decrypt wallet
	seed, err := store.OpenWallet(wallet, wp)
	if err != nil {
		return err
	}

	// Restore account handler
	key, err := mnemonic.ToPrivateKey(seed)
	if err != nil {
		return err
	}
	account, err := ac.AccountFromPrivateKey(key)
	if err != nil {
		return err
	}

	// Get client connection
	conn, err := getClientConnection()
	if err != nil {
		return fmt.Errorf("failed to establish connection: %s", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	cl := protov1.NewAgentAPIClient(conn)

	// Get transaction parameters
	txParams, err := cl.TxParameters(context.TODO(), &emptypb.Empty{})
	if err != nil {
		return err
	}
	params := at.SuggestedParams{}
	if err := json.Unmarshal(txParams.Params, &params); err != nil {
		return err
	}

	// Get sender address
	sender := account.Address.String()
	log.Debugf("%+v", params)
	log.Debug(sender, receiver, amount)

	// Build transaction
	tx, err := future.MakePaymentTxn(sender, receiver, uint64(amount), nil, "", params)
	if err != nil {
		return err
	}

	// Sign transaction
	log.Debug("signing transaction")
	_, stx, err := ac.SignTransaction(key, tx)
	if err != nil {
		return err
	}
	if !viper.GetBool("tx.submit") {
		log.Infof("signed transaction: %s", base64.StdEncoding.EncodeToString(stx))
		return nil
	}

	// Submit transaction
	log.Debug("submitting signed transaction")
	tr, err := cl.TxSubmit(context.TODO(), &protov1.TxSubmitRequest{
		Stx: stx,
	})
	if err != nil {
		return err
	}
	log.Infof("transaction successfully submitted with id: %s", tr.Id)
	return nil
}

func getTxParameters(args []string) (wallet, receiver string, amount int, err error) {
	// Wallet name
	if len(args) != 1 {
		wallet, err = readValue("enter wallet name")
		if err != nil {
			return "", "", 0, err
		}
	} else {
		wallet = args[0]
	}
	wallet = sanitize.Name(wallet)

	// Receiver address
	receiver = viper.GetString("tx.to")
	if receiver == "" {
		if receiver, err = readValue("enter receiver address"); err != nil {
			return "", "", 0, err
		}
	}

	// Tx amount
	amount = viper.GetInt("tx.amount")
	if amount == 0 {
		a, err := readValue("enter amount to send")
		if err != nil {
			return "", "", 0, err
		}
		amount, err = strconv.Atoi(a)
		if err != nil {
			return "", "", 0, err
		}
	}
	return wallet, receiver, amount, nil
}
