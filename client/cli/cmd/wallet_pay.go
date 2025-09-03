package cmd

import (
	"encoding/base64"
	"strconv"

	ac "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
)

var walletPayCmd = &cobra.Command{
	Use:     "pay",
	Short:   "Create and submit a new transaction",
	Aliases: []string{"txn", "send"},
	Example: "algoid wallet pay [wallet-name] [network]",
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
			Usage:     "Submit transaction to the network (based on your active profile)",
			FlagKey:   "tx.submit",
			ByDefault: false,
			Short:     "s",
		},
	}
	if err := cli.SetupCommandParams(walletPayCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	walletCmd.AddCommand(walletPayCmd)
}

func runWalletPayCmd(_ *cobra.Command, args []string) (err error) {
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

	// Get network client
	cl, err := getAlgoClient()
	if err != nil {
		return err
	}

	network := args[1]
	// Get transaction parameters
	params, _ := cl.Networks[network].SuggestedParams()

	// Get sender address
	sender := account.Address.String()

	// Build transaction
	tx, err := transaction.MakePaymentTxn(sender, receiver, amount, nil, "", params)
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
	txID, err := cl.Networks[network].SubmitTx(stx)
	if err != nil {
		return err
	}
	log.Infof("transaction successfully submitted with id: %s", txID)
	return nil
}

func getTxParameters(args []string) (wallet, receiver string, amount uint64, err error) {
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
	amount = viper.GetUint64("tx.amount")
	if amount == 0 {
		a, err := readValue("enter amount to send")
		if err != nil {
			return "", "", 0, err
		}
		amount, err = strconv.ParseUint(a, 10, 64)
		if err != nil {
			return "", "", 0, err
		}
	}
	return wallet, receiver, amount, nil
}
