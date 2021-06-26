package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/MDLlife/MDL/src/cipher"
	"github.com/MDLlife/MDL/src/cipher/bip39"
    "github.com/MDLlife/MDL/src/cipher/bip44"
    secp256k1 "github.com/MDLlife/MDL/src/cipher/secp256k1-go"
	"github.com/MDLlife/MDL/src/wallet"
)

const (
	// AlphaNumericSeedLength is the size of generated alphanumeric seeds, in bytes
	AlphaNumericSeedLength = 64
)

func walletCreateCmd() *cobra.Command {
	walletCreateCmd := &cobra.Command{
		Args:  cobra.ExactArgs(1),
		Use:   "walletCreate [wallet]",
		Short: "Create a new wallet",
		Long: `Create a new wallet.

    Use caution when using the "-p" command. If you have command
    history enabled your wallet encryption password can be recovered
    from the history log. If you do not include the "-p" option you will
    be prompted to enter your password after you enter your command.

    All results are returned in JSON format in addition to being written to the specified filename.`,
		SilenceUsage: true,
		RunE:         generateWalletHandler,
	}

	walletCreateCmd.Flags().BoolP("random", "r", false, "A random alpha numeric seed will be generated.")
	walletCreateCmd.Flags().BoolP("mnemonic", "m", false, "A mnemonic seed consisting of 12 dictionary words will be generated")
	walletCreateCmd.Flags().Uint64P("wordcount", "w", 12, "Number of seed words to use for mnemonic. Must be 12, 15, 18, 21 or 24")
	walletCreateCmd.Flags().StringP("seed", "s", "", "Your seed")
	walletCreateCmd.Flags().StringP("seed-passphrase", "", "", "Seed passphrase (bip44 wallets only)")
	walletCreateCmd.Flags().Uint32P("bip44-coin", "", uint32(bip44.CoinTypeMDL), "BIP44 coin type")
	walletCreateCmd.Flags().StringP("coin", "c", string(wallet.CoinTypeMDL), "Wallet address coin type (options: skycoin, bitcoin)")
	walletCreateCmd.Flags().Uint64P("num", "n", 1, `Number of addresses to generate.`)
	walletCreateCmd.Flags().StringP("label", "l", "", "Label used to identify your wallet.")
	walletCreateCmd.Flags().StringP("type", "t", wallet.WalletTypeDeterministic, "Wallet type. Types are \"collection\", \"deterministic\", \"bip44\" or \"xpub\"")
	walletCreateCmd.Flags().BoolP("encrypt", "e", false, "Create encrypted wallet.")
	walletCreateCmd.Flags().StringP("crypto-type", "x", string(wallet.DefaultCryptoType), "The crypto type for wallet encryption, can be scrypt-chacha20poly1305 or sha256-xor")
	walletCreateCmd.Flags().StringP("password", "p", "", "Wallet password")
	walletCreateCmd.Flags().StringP("xpub", "", "", "xpub key for \"xpub\" type wallets")

	return walletCreateCmd
}

func generateWalletHandler(c *cobra.Command, args []string) error {
	wltName := args[0]
	// wallet filename must have the correct extension
	if filepath.Ext(wltName) != walletExt {
		return ErrWalletName
	}

	// check if the wallet file does exist
	if _, err := os.Stat(wltName); err == nil {
		return fmt.Errorf("%v already exists", wltName)
	}

	// get number of address that are need to be generated
	num, err := c.Flags().GetUint64("num")
	if err != nil {
		return err
	}
	if num == 0 {
		return errors.New("-n must > 0")
	}

	label := c.Flag("label").Value.String()

	s := c.Flag("seed").Value.String()
	random, err := c.Flags().GetBool("random")
	if err != nil {
		return err
	}

	mnemonic, err := c.Flags().GetBool("mnemonic")
	if err != nil {
		return err
	}

	wordCount, err := c.Flags().GetUint64("wordcount")
	if err != nil {
		return err
	}

	if !mnemonic && c.Flags().Changed("wordcount") {
		return errors.New("-m must also be set when using -wordcount")
	}

	encrypt, err := c.Flags().GetBool("encrypt")
	if err != nil {
		return err
	}

	walletType, err := c.Flags().GetString("type")
	if err != nil {
		return err
	}
	if !wallet.IsValidWalletType(walletType) {
		return wallet.ErrInvalidWalletType
	}

	coinStr, err := c.Flags().GetString("coin")
	if err != nil {
		return err
	}
	coin, err := wallet.ResolveCoinType(coinStr)
	if err != nil {
		return err
	}

	var bip44Coin *bip44.CoinType
	if c.Flags().Changed("bip44-coin") {
		bip44CoinInt, err := c.Flags().GetUint32("bip44-coin")
		if err != nil {
			return err
		}

		c := bip44.CoinType(bip44CoinInt)
		bip44Coin = &c
	}

	xpub, err := c.Flags().GetString("xpub")
	if err != nil {
		return err
	}

	var sd string
	switch walletType {
	case wallet.WalletTypeBip44:
		var err error
		sd, err = parseBip44WalletSeedOptions(s, random, mnemonic, wordCount)
		if err != nil {
			return err
		}

	case wallet.WalletTypeDeterministic:
		var err error
		sd, err = parseDeterministicWalletSeedOptions(s, random, mnemonic, wordCount)
		if err != nil {
			return err
		}

	case wallet.WalletTypeCollection:
		if s != "" || random || mnemonic {
			return fmt.Errorf("%q type wallets do not use seeds", walletType)
		}
		if c.Flags().Changed("num") {
			return fmt.Errorf("%q type wallets do not support address generation", walletType)
		}
		num = 0

	case wallet.WalletTypeXPub:
		if s != "" || random || mnemonic {
			return fmt.Errorf("%q type wallets do not use seeds", walletType)
		}

	default:
		return fmt.Errorf("unhandled wallet type %q", walletType)
	}

	seedPassphrase, err := c.Flags().GetString("seed-passphrase")
	if err != nil {
		return err
	}

	cryptoType, err := wallet.CryptoTypeFromString(c.Flag("crypto-type").Value.String())
	if err != nil {
		return err
	}

	pr := NewPasswordReader([]byte(c.Flag("password").Value.String()))
	switch pr.(type) {
	case PasswordFromBytes:
		p, err := pr.Password()
		if err != nil {
			return err
		}

		if !encrypt && len(p) != 0 {
			return errors.New("password should not be set as we're not going to create a wallet with encryption")
		}
	}

	var password []byte
	if encrypt {
		var err error
		password, err = pr.Password()
		if err != nil {
			return err
		}
	}

	opts := wallet.Options{
		Label:          label,
		Seed:           sd,
		SeedPassphrase: seedPassphrase,
		Encrypt:        encrypt,
		CryptoType:     cryptoType,
		Password:       password,
		Type:           walletType,
		GenerateN:      num,
		Coin:           coin,
		Bip44Coin:      bip44Coin,
		XPub:           xpub,
	}

	wlt, err := wallet.NewWallet(filepath.Base(wltName), opts)
	if err != nil {
		return err
	}

	if err := wallet.Save(wlt, filepath.Dir(wltName)); err != nil {
		return err
	}

	return printJSON(wlt.ToReadable())
}

// wordCountToEntropy maps a mnemonic word count to its entropy size in bits
func wordCountToEntropy(wc uint64) (int, error) {
	switch wc {
	case 12:
		return 128, nil
	case 15:
		return 160, nil
	case 18:
		return 192, nil
	case 21:
		return 224, nil
	case 24:
		return 256, nil
	default:
		return 0, errors.New("word count must be 12, 15, 18, 21 or 24")
	}
}

func newMnemomic(wc uint64) (string, error) {
	entropySize, err := wordCountToEntropy(wc)
	if err != nil {
		return "", err
	}
	e, err := bip39.NewEntropy(entropySize)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(e)
}

func parseBip44WalletSeedOptions(s string, r, m bool, wc uint64) (string, error) {
	if s != "" && (r || m) {
		return "", errors.New("-r and -m can't be used with -s")
	}

	if r {
		return "", errors.New("-r can't be used for bip44 wallets")
	}

	if m || s == "" {
		var err error
		s, err = newMnemomic(wc)
		if err != nil {
			return "", err
		}
	}

	if err := bip39.ValidateMnemonic(s); err != nil {
		return "", fmt.Errorf("seed must be a valid bip39 mnemonic: %v", err)
	}

	return s, nil
}

func parseDeterministicWalletSeedOptions(s string, r, m bool, wc uint64) (string, error) {
	if s != "" {
		// 111, 101, 110
		if r || m {
			return "", errors.New("seed already specified, must not use -r or -m again")
		}
		// 100
		return s, nil
	}

	// 011
	if r && m {
		return "", errors.New("for -r and -m, only one option can be used")
	}

	// 010
	if r {
		return MakeAlphanumericSeed(), nil
	}

	// 001, 000
	return newMnemomic(wc)
}

// PUBLIC

// MakeAlphanumericSeed creates a random seed with AlphaNumericSeedLength bytes and hex encodes it
func MakeAlphanumericSeed() string {
	seedRaw := cipher.SumSHA256(secp256k1.RandByte(AlphaNumericSeedLength))
	return hex.EncodeToString(seedRaw[:])
}
