package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MDLlife/MDL/src/params"
	"github.com/MDLlife/MDL/src/readable"
	"github.com/MDLlife/MDL/src/transaction"
	"github.com/MDLlife/MDL/src/util/droplet"
	"github.com/MDLlife/MDL/src/util/fee"
	"github.com/MDLlife/MDL/src/util/mathutil"

	"github.com/MDLlife/MDL/src/cipher"
	"github.com/MDLlife/MDL/src/coin"
	"github.com/MDLlife/MDL/src/visor"
	"github.com/MDLlife/MDL/src/wallet"

	"encoding/csv"
	"encoding/json"

	"github.com/spf13/cobra"
)

var (
	// ErrTemporaryInsufficientBalance is returned if a wallet does not have enough balance for a spend, but will have enough after unconfirmed transactions confirm
	ErrTemporaryInsufficientBalance = errors.New("balance is not sufficient. Balance will be sufficient after unconfirmed transactions confirm")
)

// SendAmount represents an amount to send to an address
type SendAmount struct {
	Addr  string
	Coins uint64
}

type sendAmountJSON struct {
	Addr  string `json:"addr"`
	Coins string `json:"coins"`
}

func createRawTxnCmd() *cobra.Command {
	createRawTxnCmd := &cobra.Command{
		Short: "Create a raw transaction to be broadcast to the network later",
		Use:   "createRawTransaction [flags] [to address] [amount]",
		Long: fmt.Sprintf(`Note: The [amount] argument is the coins you will spend, 1 coins = 1e6 droplets.
    The default wallet (%s) will be used if no wallet and address was specified.

    If you are sending from a wallet the coins will be taken iteratively
    from all addresses within the wallet starting with the first address until
    the amount of the transaction is met.

    Use caution when using the "-p" command. If you have command history enabled
    your wallet encryption password can be recovered from the history log. If you
    do not include the "-p" option you will be prompted to enter your password
    after you enter your command.`, cliConfig.FullWalletPath()),
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(0),
		RunE: func(c *cobra.Command, args []string) error {
			jsonOutput, err := c.Flags().GetBool("json")
			if err != nil {
				return err
			}

			txn, err := createRawTxnCmdHandler(c, args)
			switch err.(type) {
			case nil:
			case WalletLoadError:
				printHelp(c)
				return err
			default:
				return err
			}

			rawTxn, err := txn.SerializeHex()
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(struct {
					RawTx string `json:"rawtx"`
				}{
					RawTx: rawTxn,
				})
			}

			fmt.Println(rawTxn)

			return nil
		},
	}

	createRawTxnCmd.Flags().StringP("wallet-file", "f", "", "wallet file or path. If no path is specified your default wallet path will be used.")
	createRawTxnCmd.Flags().StringP("address", "a", "", "From address")
	createRawTxnCmd.Flags().StringP("change-address", "c", "", `Specify different change address.
By default the from address or a wallets coinbase address will be used.`)
	createRawTxnCmd.Flags().StringP("many", "m", "", `use JSON string to set multiple receive addresses and coins,
example: -m '[{"addr":"$addr1", "coins": "10.2"}, {"addr":"$addr2", "coins": "20"}]'`)
	createRawTxnCmd.Flags().StringP("password", "p", "", "Wallet password")
	createRawTxnCmd.Flags().BoolP("json", "j", false, "Returns the results in JSON format.")
	createRawTxnCmd.Flags().String("csv", "", "CSV file containing addresses and amounts to send")

	return createRawTxnCmd
}

type walletAddress struct {
	Wallet  string
	Address string
}

func fromWalletOrAddress(c *cobra.Command) (walletAddress, error) {
	walletFile, err := c.Flags().GetString("wallet-file")
	if err != nil {
		return walletAddress{}, nil
	}

	address, err := c.Flags().GetString("address")
	if err != nil {
		return walletAddress{}, nil
	}

	wlt, err := resolveWalletPath(cliConfig, walletFile)
	if err != nil {
		return walletAddress{}, err
	}

	wltAddr := walletAddress{
		Wallet: wlt,
	}

	wltAddr.Address = address
	if wltAddr.Address == "" {
		return wltAddr, nil
	}

	if _, err := cipher.DecodeBase58Address(wltAddr.Address); err != nil {
		return walletAddress{}, fmt.Errorf("invalid address: %s", wltAddr.Address)
	}

	return wltAddr, nil
}

func getChangeAddress(wltAddr walletAddress, chgAddr string) (string, error) {
	if chgAddr == "" {
		switch {
		case wltAddr.Address != "":
			// use the from address as change address
			chgAddr = wltAddr.Address
		case wltAddr.Wallet != "":
			// get the default wallet's coin base address
			wlt, err := wallet.Load(wltAddr.Wallet)
			if err != nil {
				return "", WalletLoadError{err}
			}

			if len(wlt.Entries) > 0 {
				chgAddr = wlt.Entries[0].Address.String()
			} else {
				return "", errors.New("no change address was found")
			}
		default:
			return "", errors.New("both wallet file, from address and change address are empty")
		}
	}

	// validate the address
	_, err := cipher.DecodeBase58Address(chgAddr)
	if err != nil {
		return "", fmt.Errorf("invalid change address: %s", chgAddr)
	}

	return chgAddr, nil
}

func getToAddresses(c *cobra.Command, args []string) ([]SendAmount, error) {
	csvFile, err := c.Flags().GetString("csv")
	if err != nil {
		return nil, err
	}
	many, err := c.Flags().GetString("many")
	if err != nil {
		return nil, err
	}

	if csvFile != "" && many != "" {
		return nil, errors.New("-csv and -m cannot be combined")
	}

	if many != "" {
		return parseSendAmountsFromJSON(many)
	} else if csvFile != "" {
		fields, err := openCSV(csvFile)
		if err != nil {
			return nil, err
		}
		return parseSendAmountsFromCSV(fields)
	}

	if len(args) < 2 {
		return nil, fmt.Errorf("requires at least 2 arg(s), only received %d", len(args))
	}

	toAddr := args[0]

	if _, err := cipher.DecodeBase58Address(toAddr); err != nil {
		return nil, err
	}

	amt, err := getAmount(args)
	if err != nil {
		return nil, err
	}

	return []SendAmount{{
		Addr:  toAddr,
		Coins: amt,
	}}, nil
}

func openCSV(csvFile string) ([][]string, error) {
	f, err := os.Open(csvFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	return r.ReadAll()
}

func parseSendAmountsFromCSV(fields [][]string) ([]SendAmount, error) {
	var sends []SendAmount
	var errs []error
	for i, f := range fields {
		addr := f[0]

		addr = strings.TrimSpace(addr)

		if _, err := cipher.DecodeBase58Address(addr); err != nil {
			err = fmt.Errorf("[row %d] Invalid address %s: %v", i, addr, err)
			errs = append(errs, err)
			continue
		}

		coins, err := droplet.FromString(f[1])
		if err != nil {
			err = fmt.Errorf("[row %d] Invalid amount %s: %v", i, f[1], err)
			errs = append(errs, err)
			continue
		}

		sends = append(sends, SendAmount{
			Addr:  addr,
			Coins: coins,
		})
	}

	if len(errs) > 0 {
		errMsgs := make([]string, len(errs))
		for i, err := range errs {
			errMsgs[i] = err.Error()
		}

		errMsg := strings.Join(errMsgs, "\n")

		return nil, errors.New(errMsg)
	}

	return sends, nil
}

func parseSendAmountsFromJSON(m string) ([]SendAmount, error) {
	sas := []sendAmountJSON{}

	if err := json.NewDecoder(strings.NewReader(m)).Decode(&sas); err != nil {
		return nil, fmt.Errorf("invalid -m flag string, err: %v", err)
	}

	sendAmts := make([]SendAmount, 0, len(sas))

	for _, sa := range sas {
		amt, err := droplet.FromString(sa.Coins)
		if err != nil {
			return nil, fmt.Errorf("invalid coins value in -m flag string: %v", err)
		}

		sendAmts = append(sendAmts, SendAmount{
			Addr:  sa.Addr,
			Coins: amt,
		})
	}

	return sendAmts, nil
}

func getAmount(args []string) (uint64, error) {
	amount := args[1]
	amt, err := droplet.FromString(amount)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %v", err)
	}

	return amt, nil
}

// createRawTxnArgs are encapsulated arguments for creating a transaction
type createRawTxnArgs struct {
	WalletID      string
	Address       string
	ChangeAddress string
	SendAmounts   []SendAmount
	Password      PasswordReader
}

func parseCreateRawTxnArgs(c *cobra.Command, args []string) (*createRawTxnArgs, error) {
	wltAddr, err := fromWalletOrAddress(c)
	if err != nil {
		return nil, err
	}

	changeAddress, err := c.Flags().GetString("change-address")
	if err != nil {
		return nil, err
	}
	chgAddr, err := getChangeAddress(wltAddr, changeAddress)
	if err != nil {
		return nil, err
	}

	toAddrs, err := getToAddresses(c, args)
	if err != nil {
		return nil, err
	}
	if err := validateSendAmounts(toAddrs); err != nil {
		return nil, err
	}

	password, err := c.Flags().GetString("password")
	if err != nil {
		return nil, err
	}
	pr := NewPasswordReader([]byte(password))

	return &createRawTxnArgs{
		WalletID:      wltAddr.Wallet,
		Address:       wltAddr.Address,
		ChangeAddress: chgAddr,
		SendAmounts:   toAddrs,
		Password:      pr,
	}, nil
}

func createRawTxnCmdHandler(c *cobra.Command, args []string) (*coin.Transaction, error) {
	parsedArgs, err := parseCreateRawTxnArgs(c, args)
	if err != nil {
		return nil, err
	}

	if parsedArgs.Address == "" {
		return CreateRawTxnFromWallet(apiClient, parsedArgs.WalletID, parsedArgs.ChangeAddress, parsedArgs.SendAmounts, parsedArgs.Password)
	}

	return CreateRawTxnFromAddress(apiClient, parsedArgs.Address, parsedArgs.WalletID, parsedArgs.ChangeAddress, parsedArgs.SendAmounts, parsedArgs.Password)
}

func validateSendAmounts(toAddrs []SendAmount) error {
	for _, arg := range toAddrs {
		// validate to address
		_, err := cipher.DecodeBase58Address(arg.Addr)
		if err != nil {
			return ErrAddress
		}

		if arg.Coins == 0 {
			return errors.New("Cannot send 0 coins")
		}
	}

	if len(toAddrs) == 0 {
		return errors.New("No destination addresses")
	}

	return nil
}

// PUBLIC

// CreateRawTxnFromWallet creates a transaction from any address or combination of addresses in a wallet
func CreateRawTxnFromWallet(c GetOutputser, walletFile, chgAddr string, toAddrs []SendAmount, pr PasswordReader) (*coin.Transaction, error) {
	// check change address
	cAddr, err := cipher.DecodeBase58Address(chgAddr)
	if err != nil {
		return nil, ErrAddress
	}

	// check if the change address is in wallet.
	wlt, err := wallet.Load(walletFile)
	if err != nil {
		return nil, err
	}

	_, ok := wlt.GetEntry(cAddr)
	if !ok {
		return nil, fmt.Errorf("change address %v is not in wallet", chgAddr)
	}

	switch pr.(type) {
	case nil:
		if wlt.IsEncrypted() {
			return nil, wallet.ErrWalletEncrypted
		}
	case PasswordFromBytes:
		p, err := pr.Password()
		if err != nil {
			return nil, err
		}

		if !wlt.IsEncrypted() && len(p) != 0 {
			return nil, wallet.ErrWalletNotEncrypted
		}
	}

	var password []byte
	if wlt.IsEncrypted() {
		var err error
		password, err = pr.Password()
		if err != nil {
			return nil, err
		}
	}

	// get all address in the wallet
	totalAddrs := wlt.GetAddresses()
	addrStrArray := make([]string, len(totalAddrs))
	for i, a := range totalAddrs {
		addrStrArray[i] = a.String()
	}

	return CreateRawTxn(c, wlt, addrStrArray, chgAddr, toAddrs, password)
}

// CreateRawTxnFromAddress creates a transaction from a specific address in a wallet
func CreateRawTxnFromAddress(c GetOutputser, addr, walletFile, chgAddr string, toAddrs []SendAmount, pr PasswordReader) (*coin.Transaction, error) {
	// check if the address is in the default wallet.
	wlt, err := wallet.Load(walletFile)
	if err != nil {
		return nil, err
	}

	srcAddr, err := cipher.DecodeBase58Address(addr)
	if err != nil {
		return nil, ErrAddress
	}

	_, ok := wlt.GetEntry(srcAddr)
	if !ok {
		return nil, fmt.Errorf("%v address is not in wallet", addr)
	}

	// validate change address
	cAddr, err := cipher.DecodeBase58Address(chgAddr)
	if err != nil {
		return nil, ErrAddress
	}

	_, ok = wlt.GetEntry(cAddr)
	if !ok {
		return nil, fmt.Errorf("change address %v is not in wallet", chgAddr)
	}

	switch pr.(type) {
	case nil:
		if wlt.IsEncrypted() {
			return nil, wallet.ErrWalletEncrypted
		}
	case PasswordFromBytes:
		p, err := pr.Password()
		if err != nil {
			return nil, err
		}

		if !wlt.IsEncrypted() && len(p) != 0 {
			return nil, wallet.ErrWalletNotEncrypted
		}
	}

	var password []byte
	if wlt.IsEncrypted() {
		var err error
		password, err = pr.Password()
		if err != nil {
			return nil, err
		}
	}

	return CreateRawTxn(c, wlt, []string{addr}, chgAddr, toAddrs, password)
}

// GetOutputser implements unspent output querying
type GetOutputser interface {
	OutputsForAddresses([]string) (*readable.UnspentOutputsSummary, error)
}

// CreateRawTxn creates a transaction from a set of addresses contained in a loaded *wallet.Wallet
func CreateRawTxn(c GetOutputser, wlt *wallet.Wallet, inAddrs []string, chgAddr string, toAddrs []SendAmount, password []byte) (*coin.Transaction, error) {
	if err := validateSendAmounts(toAddrs); err != nil {
		return nil, err
	}

	// Get unspent outputs of those addresses
	outputs, err := c.OutputsForAddresses(inAddrs)
	if err != nil {
		return nil, err
	}

	inUxs, err := outputs.SpendableOutputs().ToUxArray()
	if err != nil {
		return nil, err
	}

	txn, err := createRawTxn(outputs, wlt, chgAddr, toAddrs, password)
	if err != nil {
		return nil, err
	}

	// filter out unspents which are not used in transaction
	var inUxsFiltered coin.UxArray
	for _, h := range txn.In {
		for _, u := range inUxs {
			if h == u.Hash() {
				inUxsFiltered = append(inUxsFiltered, u)
			}
		}
	}

	head, err := outputs.Head.ToCoinBlockHeader()
	if err != nil {
		return nil, err
	}

	if err := visor.VerifySingleTxnSoftConstraints(*txn, head.Time, inUxsFiltered, params.UserVerifyTxn); err != nil {
		return nil, err
	}
	if err := visor.VerifySingleTxnHardConstraints(*txn, head, inUxsFiltered, visor.TxnSigned); err != nil {
		return nil, err
	}
	if err := visor.VerifySingleTxnUserConstraints(*txn); err != nil {
		return nil, err
	}

	return txn, nil
}

func createRawTxn(uxouts *readable.UnspentOutputsSummary, wlt *wallet.Wallet, chgAddr string, toAddrs []SendAmount, password []byte) (*coin.Transaction, error) {
	// Calculate total required coins
	var totalCoins uint64
	for _, arg := range toAddrs {
		var err error
		totalCoins, err = mathutil.AddUint64(totalCoins, arg.Coins)
		if err != nil {
			return nil, err
		}
	}

	spendOutputs, err := chooseSpends(uxouts, totalCoins)
	if err != nil {
		return nil, err
	}

	txOuts, err := makeChangeOut(spendOutputs, chgAddr, toAddrs)
	if err != nil {
		return nil, err
	}

	f := func(w *wallet.Wallet) (*coin.Transaction, error) {
		keys, err := getKeys(w, spendOutputs)
		if err != nil {
			return nil, err
		}

		return NewTransaction(spendOutputs, keys, txOuts)
	}

	makeTxn := func() (*coin.Transaction, error) {
		return f(wlt)
	}

	if wlt.IsEncrypted() {
		makeTxn = func() (*coin.Transaction, error) {
			var tx *coin.Transaction
			if err := wlt.GuardView(password, func(w *wallet.Wallet) error {
				var err error
				tx, err = f(w)
				return err
			}); err != nil {
				return nil, err
			}

			return tx, nil
		}
	}

	return makeTxn()
}

func chooseSpends(uxouts *readable.UnspentOutputsSummary, coins uint64) ([]transaction.UxBalance, error) {
	// Convert spendable unspent outputs to []transaction.UxBalance
	spendableOutputs, err := readable.OutputsToUxBalances(uxouts.SpendableOutputs())
	if err != nil {
		return nil, err
	}

	// Choose which unspent outputs to spend
	// Use the MinimizeUxOuts strategy, since this is most likely used by
	// application that may need to send frequently.
	// Using fewer UxOuts will leave more available for other transactions,
	// instead of waiting for confirmation.
	outs, err := transaction.ChooseSpendsMinimizeUxOuts(spendableOutputs, coins, 0)
	if err != nil {
		// If there is not enough balance in the spendable outputs,
		// see if there is enough balance when including incoming outputs
		if err == transaction.ErrInsufficientBalance {
			expectedOutputs, otherErr := readable.OutputsToUxBalances(uxouts.ExpectedOutputs())
			if otherErr != nil {
				return nil, otherErr
			}

			if _, otherErr := transaction.ChooseSpendsMinimizeUxOuts(expectedOutputs, coins, 0); otherErr != nil {
				return nil, err
			}

			return nil, ErrTemporaryInsufficientBalance
		}

		return nil, err
	}

	return outs, nil
}

func makeChangeOut(outs []transaction.UxBalance, chgAddr string, toAddrs []SendAmount) ([]coin.TransactionOutput, error) {
	var totalInCoins, totalInHours, totalOutCoins uint64

	for _, o := range outs {
		totalInCoins += o.Coins
		totalInHours += o.Hours
	}

	if totalInHours == 0 {
		return nil, fee.ErrTxnNoFee
	}

	for _, to := range toAddrs {
		totalOutCoins += to.Coins
	}

	if totalInCoins < totalOutCoins {
		return nil, transaction.ErrInsufficientBalance
	}

	outAddrs := []coin.TransactionOutput{}
	changeAmount := totalInCoins - totalOutCoins

	haveChange := changeAmount > 0
	nAddrs := uint64(len(toAddrs))
	changeHours, addrHours, totalOutHours := transaction.DistributeSpendHours(totalInHours, nAddrs, haveChange)

	if err := fee.VerifyTransactionFeeForHours(totalOutHours, totalInHours-totalOutHours, params.UserVerifyTxn.BurnFactor); err != nil {
		return nil, err
	}

	for i, to := range toAddrs {
		// check if changeHours > 0, we do not need to cap addrHours when changeHours is zero
		// changeHours is zero when there is no change left or all the coinhours were used in fees
		// 1) if there is no change then the remaining coinhours are evenly distributed among the destination addresses
		// 2) if all the coinhours are burned in fees then all addrHours are zero by default
		if changeHours > 0 {
			// the coinhours are capped to a maximum of incoming coins for the address
			// if incoming coins < 1 then the cap is set to 1 coinhour

			spendCoinsAmt := to.Coins / 1e6
			if spendCoinsAmt == 0 {
				spendCoinsAmt = 1
			}

			// allow addrHours to be less than the incoming coins of the address but not more
			if addrHours[i] > spendCoinsAmt {
				// cap the addrHours, move the difference to changeHours
				changeHours += addrHours[i] - spendCoinsAmt
				addrHours[i] = spendCoinsAmt
			}
		}

		outAddrs = append(outAddrs, mustMakeUtxoOutput(to.Addr, to.Coins, addrHours[i]))
	}

	if haveChange {
		outAddrs = append(outAddrs, mustMakeUtxoOutput(chgAddr, changeAmount, changeHours))
	}

	return outAddrs, nil
}

func mustMakeUtxoOutput(addr string, coins, hours uint64) coin.TransactionOutput {
	uo := coin.TransactionOutput{}
	uo.Address = cipher.MustDecodeBase58Address(addr)
	uo.Coins = coins
	uo.Hours = hours
	return uo
}

func getKeys(wlt *wallet.Wallet, outs []transaction.UxBalance) ([]cipher.SecKey, error) {
	keys := make([]cipher.SecKey, len(outs))
	for i, o := range outs {
		entry, ok := wlt.GetEntry(o.Address)
		if !ok {
			return nil, fmt.Errorf("%v is not in wallet", o.Address.String())
		}

		keys[i] = entry.Secret
	}
	return keys, nil
}

// NewTransaction creates a transaction. The transaction should be validated against hard and soft constraints before transmission.
func NewTransaction(utxos []transaction.UxBalance, keys []cipher.SecKey, outs []coin.TransactionOutput) (*coin.Transaction, error) {
	txn := coin.Transaction{}
	for _, u := range utxos {
		if err := txn.PushInput(u.Hash); err != nil {
			return nil, err
		}
	}

	for _, o := range outs {
		if err := txn.PushOutput(o.Address, o.Coins, o.Hours); err != nil {
			return nil, err
		}
	}

	txn.SignInputs(keys)

	err := txn.UpdateHeader()
	if err != nil {
		return nil, err
	}

	return &txn, nil
}
