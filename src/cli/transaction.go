package cli

import (
	"errors"
	"fmt"

	"github.com/MDLlife/MDL/src/api"
	"github.com/MDLlife/MDL/src/cipher"
	"github.com/MDLlife/MDL/src/coin"
	"github.com/MDLlife/MDL/src/readable"

	"github.com/spf13/cobra"
)

// TxnResult wraps readable.TransactionWithStatus
type TxnResult struct {
	Transaction *readable.TransactionWithStatus `json:"transaction"`
}

func transactionCmd() *cobra.Command {
	return &cobra.Command{
		Short:                 "Show detail info of specific transaction",
		Use:                   "transaction [transaction id]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			txid := args[0]
			if txid == "" {
				return errors.New("txid is empty")
			}

			// validate the txid
			_, err := cipher.SHA256FromHex(txid)
			if err != nil {
				return errors.New("invalid txid")
			}

			txn, err := apiClient.Transaction(txid)
			if err != nil {
				return err
			}

			return printJSON(TxnResult{
				Transaction: txn,
			})
		},
	}
}

func decodeRawTxnCmd() *cobra.Command {
	return &cobra.Command{
		Short:                 "Decode raw transaction",
		Use:                   "decodeRawTransaction [raw transaction]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			txn, err := coin.DeserializeTransactionHex(args[0])
			if err != nil {
				return fmt.Errorf("invalid raw transaction: %v", err)
			}

			// Assume the transaction is not malformed and if it has no inputs
			// that it is the genesis block's transaction
			isGenesis := len(txn.In) == 0
			rTxn, err := readable.NewTransaction(txn, isGenesis)
			if err != nil {
				return err
			}

			return printJSON(rTxn)
		},
	}
}

func addressTransactionsCmd() *cobra.Command {
	return &cobra.Command{
		Short: "Show detail for transaction associated with one or more specified addresses",
		Use:   "addressTransactions [address list]",
		Long: `Display transactions for specific addresses, seperate multiple addresses with a space,
        example: addressTransactions addr1 addr2 addr3`,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		RunE:                  getAddressTransactionsCmd,
	}
}

func getAddressTransactionsCmd(c *cobra.Command, args []string) error {
	// Build the list of addresses from the command line arguments
	addrs := make([]string, len(args))
	var err error
	for i := 0; i < len(args); i++ {
		addrs[i] = args[i]
		if _, err = cipher.DecodeBase58Address(addrs[i]); err != nil {
			return fmt.Errorf("invalid address: %v, err: %v", addrs[i], err)
		}
	}

	// If one or more addresses have been provided, request their transactions - otherwise report an error
	if len(addrs) > 0 {
		outputs, err := apiClient.TransactionsVerbose(addrs)
		if err != nil {
			return err
		}

		return printJSON(outputs)
	}

	return fmt.Errorf("at least one address must be specified. Example: %s addr1 addr2 addr3", c.Name())
}

func verifyTransactionCmd() *cobra.Command {
	return &cobra.Command{
		Short:                 "Verify if the specific transaction is spendable",
		Use:                   "verifyTransaction [encoded transaction]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			encodedTxn := args[0]
			if encodedTxn == "" {
				return errors.New("transaction is empty")
			}

			_, err := apiClient.VerifyTransaction(api.VerifyTransactionRequest{
				EncodedTransaction: encodedTxn,
			})
			if err != nil {
				return err
			}

			fmt.Println("transaction is spendable")

			return nil
		},
	}
}

func pendingTransactionsCmd() *cobra.Command {
	pendingTxnsCmd := &cobra.Command{
		Short:                 "Get all unconfirmed transactions",
		Use:                   "pendingTransactions",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			isVerbose, err := c.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			if isVerbose {
				pendingTxns, err := apiClient.PendingTransactionsVerbose()
				if err != nil {
					return err
				}

				return printJSON(pendingTxns)
			}

			pendingTxns, err := apiClient.PendingTransactions()
			if err != nil {
				return err
			}

			return printJSON(pendingTxns)
		},
	}

	pendingTxnsCmd.Flags().BoolP("verbose", "v", false,
		`Require the transaction inputs to include the owner address, coins, hours and calculated hours.
	The hours are the original hours the output was created with.
	The calculated hours are calculated based upon the current system time, and provide an approximate 
	coin hour value of the output if it were to be confirmed at that instant.`)

	return pendingTxnsCmd
}
