/*
Package wallet implements wallets and the wallet database service
*/
package wallet

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MDLlife/MDL/src/cipher"
	"github.com/MDLlife/MDL/src/cipher/bip44"
	"github.com/MDLlife/MDL/src/util/file"
	"github.com/MDLlife/MDL/src/util/logging"
)

// Error wraps wallet-related errors.
// It wraps errors caused by user input, but not errors caused by programmer input or internal issues.
type Error struct {
	error
}

// NewError creates an Error
func NewError(err error) error {
	if err == nil {
		return nil
	}
	return Error{err}
}

var (
	// Version represents the current wallet version
	Version = "0.4"

	logger = logging.MustGetLogger("wallet")

	// ErrInvalidEncryptedField is returned if a wallet's Meta.encrypted value is invalid.
	ErrInvalidEncryptedField = NewError(errors.New(`encrypted field value is not valid, must be "true", "false" or ""`))
	// ErrWalletEncrypted is returned when trying to generate addresses or sign tx in encrypted wallet
	ErrWalletEncrypted = NewError(errors.New("wallet is encrypted"))
	// ErrWalletNotEncrypted is returned when trying to decrypt unencrypted wallet
	ErrWalletNotEncrypted = NewError(errors.New("wallet is not encrypted"))
	// ErrMissingPassword is returned when trying to create wallet with encryption, but password is not provided.
	ErrMissingPassword = NewError(errors.New("missing password"))
	// ErrMissingEncrypt is returned when trying to create wallet with password, but options.Encrypt is not set.
	ErrMissingEncrypt = NewError(errors.New("missing encrypt"))
	// ErrInvalidPassword is returned if decrypts secrets failed
	ErrInvalidPassword = NewError(errors.New("invalid password"))
	// ErrMissingSeed is returned when trying to create wallet without a seed
	ErrMissingSeed = NewError(errors.New("missing seed"))
	// ErrMissingAuthenticated is returned if try to decrypt a scrypt chacha20poly1305 encrypted wallet, and find no authenticated metadata.
	ErrMissingAuthenticated = NewError(errors.New("missing authenticated metadata"))
	// ErrWrongCryptoType is returned when decrypting wallet with wrong crypto method
	ErrWrongCryptoType = NewError(errors.New("wrong crypto type"))
	// ErrWalletNotExist is returned if a wallet does not exist
	ErrWalletNotExist = NewError(errors.New("wallet doesn't exist"))
	// ErrSeedUsed is returned if a wallet already exists with the same seed
	ErrSeedUsed = NewError(errors.New("a wallet already exists with this seed"))
	// ErrXPubKeyUsed is returned if a wallet already exists with the same xpub key
	ErrXPubKeyUsed = NewError(errors.New("a wallet already exists with this xpub key"))
	// ErrWalletAPIDisabled is returned when trying to do wallet actions while the EnableWalletAPI option is false
	ErrWalletAPIDisabled = NewError(errors.New("wallet api is disabled"))
	// ErrSeedAPIDisabled is returned when trying to get seed of wallet while the EnableWalletAPI or EnableSeedAPI is false
	ErrSeedAPIDisabled = NewError(errors.New("wallet seed api is disabled"))
	// ErrWalletNameConflict represents the wallet name conflict error
	ErrWalletNameConflict = NewError(errors.New("wallet name would conflict with existing wallet, renaming"))
	// ErrWalletRecoverSeedWrong is returned if the seed or seed passphrase does not match the specified wallet when recovering
	ErrWalletRecoverSeedWrong = NewError(errors.New("wallet recovery seed or seed passphrase is wrong"))
	// ErrNilTransactionsFinder is returned if Options.ScanN > 0 but a nil TransactionsFinder was provided
	ErrNilTransactionsFinder = NewError(errors.New("scan ahead requested but balance getter is nil"))
	// ErrInvalidCoinType is returned for invalid coin types
	ErrInvalidCoinType = NewError(errors.New("invalid coin type"))
	// ErrInvalidWalletType is returned for invalid wallet types
	ErrInvalidWalletType = NewError(errors.New("invalid wallet type"))
	// ErrWalletTypeNotRecoverable is returned by RecoverWallet is the wallet type does not support recovery
	ErrWalletTypeNotRecoverable = NewError(errors.New("wallet type is not recoverable"))
	// ErrWalletPermission is returned when updating a wallet without writing permission
	ErrWalletPermission = NewError(errors.New("saving wallet permission denied"))
)

const (
	// WalletExt wallet file extension
	WalletExt = "wlt"

	// WalletTimestampFormat wallet timestamp layout
	WalletTimestampFormat = "2006_01_02"

	// CoinTypeMDL mdl type
	CoinTypeMDL CoinType = "mdl"

	// CoinTypeBitcoin bitcoin type
	CoinTypeBitcoin CoinType = "bitcoin"

	// WalletTypeDeterministic deterministic wallet type.
	// Uses the original Skycoin deterministic key generator.
	WalletTypeDeterministic = "deterministic"
	// WalletTypeCollection collection wallet type.
	// Does not use any key generator; keys must be added explicitly
	WalletTypeCollection = "collection"
	// WalletTypeBip44 bip44 HD wallet type.
	// Follow the bip44 spec.
	WalletTypeBip44 = "bip44"
	// WalletTypeXPub xpub HD wallet type.
	// Allows generating addresses without a secret key
	WalletTypeXPub = "xpub"
)

// ResolveCoinType normalizes a coin type string to a CoinType constant
func ResolveCoinType(s string) (CoinType, error) {
	switch strings.ToLower(s) {
	case "MDL", "mdl":
		return CoinTypeMDL, nil
	case "btc", "bitcoin":
		return CoinTypeBitcoin, nil
	default:
		return CoinType(""), ErrInvalidCoinType
	}
}

// IsValidWalletType returns true if a wallet type is recognized
func IsValidWalletType(t string) bool {
	switch t {
	case WalletTypeDeterministic,
		WalletTypeCollection,
		WalletTypeBip44,
		WalletTypeXPub:
		return true
	default:
		return false
	}
}

// CoinType represents the wallet coin type, which refers to the pubkey2addr method used
type CoinType string

// NewWalletFilename generates a filename from the current time and random bytes
func NewWalletFilename() string {
	timestamp := time.Now().Format(WalletTimestampFormat)
	// should read in wallet files and make sure does not exist
	padding := hex.EncodeToString((cipher.RandByte(2)))
	return fmt.Sprintf("%s_%s.%s", timestamp, padding, WalletExt)
}

// Options options that could be used when creating a wallet
type Options struct {
	Type           string          // wallet type: deterministic, collection. Refers to which key generation mechanism is used.
	Coin           CoinType        // coin type: skycoin, bitcoin, etc. Refers to which pubkey2addr method is used.
	Bip44Coin      *bip44.CoinType // bip44 path coin type
	Label          string          // wallet label
	Seed           string          // wallet seed
	SeedPassphrase string          // wallet seed passphrase (bip44 wallets only)
	Encrypt        bool            // whether the wallet need to be encrypted.
	Password       []byte          // password that would be used for encryption, and would only be used when 'Encrypt' is true.
	CryptoType     CryptoType      // wallet encryption type, scrypt-chacha20poly1305 or sha256-xor.
	ScanN          uint64          // number of addresses that're going to be scanned for a balance. The highest address with a balance will be used.
	GenerateN      uint64          // number of addresses to generate, regardless of balance
	XPub           string          // xpub key (xpub wallets only)
}

// newWallet creates a wallet instance with given name and options.
func newWallet(wltName string, opts Options, tf TransactionsFinder) (Wallet, error) {
	wltType := opts.Type
	if wltType == "" {
		return nil, NewError(errors.New("wallet type is required"))
	}
	if !IsValidWalletType(wltType) {
		return nil, ErrInvalidWalletType
	}

	lastSeed := ""
	if wltType == WalletTypeDeterministic {
		lastSeed = opts.Seed
	}

	var bip44Coin bip44.CoinType
	if wltType == WalletTypeBip44 {
		if opts.Bip44Coin == nil {
			switch opts.Coin {
			case CoinTypeBitcoin:
				bip44Coin = bip44.CoinTypeBitcoin
			case CoinTypeMDL:
				bip44Coin = bip44.CoinTypeMDL
			default:
				bip44Coin = bip44.CoinTypeMDL
			}
		} else {
			bip44Coin = *opts.Bip44Coin
		}
	}

	if opts.SeedPassphrase != "" && wltType != WalletTypeBip44 {
		return nil, NewError(fmt.Errorf("seedPassphrase is only used for %q wallets", WalletTypeBip44))
	}

	if opts.XPub != "" && wltType != WalletTypeXPub {
		return nil, NewError(fmt.Errorf("xpub is only used for %q wallets", WalletTypeXPub))
	}

	switch wltType {
	case WalletTypeDeterministic, WalletTypeBip44:
		if opts.Seed == "" {
			return nil, ErrMissingSeed
		}

		if opts.ScanN > 0 && tf == nil {
			return nil, ErrNilTransactionsFinder
		}

	case WalletTypeXPub:
		if opts.Seed != "" {
			return nil, NewError(fmt.Errorf("seed should not be provided for %q wallets", wltType))
		}

		if opts.ScanN > 0 && tf == nil {
			return nil, ErrNilTransactionsFinder
		}

	case WalletTypeCollection:
		if opts.Seed != "" {
			return nil, NewError(fmt.Errorf("seed should not be provided for %q wallets", wltType))
		}

	default:
		return nil, ErrInvalidWalletType
	}

	coin := opts.Coin
	if coin == "" {
		coin = CoinTypeMDL
	}
	coin, err := ResolveCoinType(string(coin))
	if err != nil {
		return nil, err
	}

	meta := Meta{
		metaFilename:       wltName,
		metaVersion:        Version,
		metaLabel:          opts.Label,
		metaSeed:           opts.Seed,
		metaLastSeed:       lastSeed,
		metaSeedPassphrase: opts.SeedPassphrase,
		metaTimestamp:      strconv.FormatInt(time.Now().Unix(), 10),
		metaType:           wltType,
		metaCoin:           string(coin),
		metaEncrypted:      "false",
		metaCryptoType:     "",
		metaSecrets:        "",
		metaXPub:           opts.XPub,
	}

	// Create the wallet
	var w Wallet
	switch wltType {
	case WalletTypeDeterministic:
		w, err = newDeterministicWallet(meta)
	case WalletTypeCollection:
		w, err = newCollectionWallet(meta)
	case WalletTypeBip44:
		meta.setBip44Coin(bip44Coin)
		w, err = newBip44Wallet(meta)
	case WalletTypeXPub:
		meta.setXPub(opts.XPub)
		w, err = newXPubWallet(meta)
	default:
		logger.Panic("unhandled wltType")
	}

	if err != nil {
		logger.WithError(err).WithField("walletType", wltType).Error("newWallet failed")
		return nil, err
	}

	// Generate wallet addresses
	switch wltType {
	case WalletTypeDeterministic, WalletTypeBip44, WalletTypeXPub:
		generateN := opts.GenerateN
		if generateN == 0 {
			generateN = 1
		}

		logger.WithFields(logrus.Fields{
			"generateN":  generateN,
			"walletType": wltType,
		}).Infof("Generating addresses for wallet")

		if _, err := w.GenerateAddresses(generateN); err != nil {
			return nil, err
		}

		if opts.ScanN != 0 && coin != CoinTypeMDL {
			return nil, errors.New("Wallet scanning is only supported for MDL address wallets")
		}

		if opts.ScanN > generateN {
			// Scan for addresses with balances
			logger.WithFields(logrus.Fields{
				"scanN":      opts.ScanN,
				"walletType": wltType,
			}).Info("Scanning addresses for wallet")
			if err := w.ScanAddresses(opts.ScanN, tf); err != nil {
				return nil, err
			}
		}

	case WalletTypeCollection:
		if opts.GenerateN != 0 || opts.ScanN != 0 {
			return nil, NewError(fmt.Errorf("wallet scanning is not defined for %q wallets", wltType))
		}

	default:
		logger.Panic("unhandled wltType")
	}

	// Validate the wallet, before encrypting
	if err := w.Validate(); err != nil {
		return nil, err
	}

	// Check if the wallet should be encrypted
	if !opts.Encrypt {
		if len(opts.Password) != 0 {
			return nil, ErrMissingEncrypt
		}
		return w, nil
	}

	// Check if the password is provided
	if len(opts.Password) == 0 {
		return nil, ErrMissingPassword
	}

	// Check crypto type
	if opts.CryptoType == "" {
		opts.CryptoType = DefaultCryptoType
	}

	if _, err := getCrypto(opts.CryptoType); err != nil {
		return nil, err
	}

	// Encrypt the wallet
	if err := Lock(w, opts.Password, opts.CryptoType); err != nil {
		return nil, err
	}

	// Validate the wallet again, after encrypting
	if err := w.Validate(); err != nil {
		return nil, err
	}

	return w, nil
}

// NewWallet creates wallet without scanning addresses
func NewWallet(wltName string, opts Options) (Wallet, error) {
	return newWallet(wltName, opts, nil)
}

// NewWalletScanAhead creates wallet and scan ahead N addresses
func NewWalletScanAhead(wltName string, opts Options, tf TransactionsFinder) (Wallet, error) {
	return newWallet(wltName, opts, tf)
}

// Lock encrypts the wallet with the given password and specific crypto type
func Lock(w Wallet, password []byte, cryptoType CryptoType) error {
	if len(password) == 0 {
		return ErrMissingPassword
	}

	if w.IsEncrypted() {
		return ErrWalletEncrypted
	}

	wlt := w.Clone()

	// Records seeds in secrets
	ss := make(Secrets)
	defer func() {
		// Wipes all unencrypted sensitive data
		ss.erase()
		wlt.Erase()
	}()

	wlt.PackSecrets(ss)

	sb, err := ss.serialize()
	if err != nil {
		return err
	}

	crypto, err := getCrypto(cryptoType)
	if err != nil {
		return err
	}

	// Encrypts the secrets
	encSecret, err := crypto.Encrypt(sb, password)
	if err != nil {
		return err
	}

	// Sets wallet as encrypted
	wlt.SetEncrypted(cryptoType, string(encSecret))

	// Update the wallet to the latest version, which indicates encryption support
	wlt.SetVersion(Version)

	// Wipes unencrypted sensitive data
	wlt.Erase()

	// Wipes the secret fields in w
	w.Erase()

	// Replace the original wallet with new encrypted wallet
	w.CopyFrom(wlt)
	return nil
}

// Unlock decrypts the wallet into a temporary decrypted copy of the wallet
// Returns error if the decryption fails
// The temporary decrypted wallet should be erased from memory when done.
func Unlock(w Wallet, password []byte) (Wallet, error) {
	if !w.IsEncrypted() {
		return nil, ErrWalletNotEncrypted
	}

	if len(password) == 0 {
		return nil, ErrMissingPassword
	}

	wlt := w.Clone()

	// Gets the secrets string
	sstr := w.Secrets()
	if sstr == "" {
		return nil, errors.New("secrets missing from wallet")
	}

	ct := w.CryptoType()
	if ct == "" {
		return nil, errors.New("missing crypto type")
	}

	// Gets the crypto module
	crypto, err := getCrypto(ct)
	if err != nil {
		return nil, err
	}

	// Decrypts the secrets
	sb, err := crypto.Decrypt([]byte(sstr), password)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	defer func() {
		// Wipe the data from the secrets bytes buffer
		for i := range sb {
			sb[i] = 0
		}
	}()

	// Deserialize into secrets
	ss := make(Secrets)
	defer ss.erase()
	if err := ss.deserialize(sb); err != nil {
		return nil, err
	}

	if err := wlt.UnpackSecrets(ss); err != nil {
		return nil, err
	}

	wlt.SetDecrypted()

	return wlt, nil
}

// Wallet defines the wallet API
type Wallet interface {
	Find(string) string
	Seed() string
	LastSeed() string
	SeedPassphrase() string
	Timestamp() int64
	SetTimestamp(int64)
	Coin() CoinType
	Bip44Coin() bip44.CoinType
	Type() string
	Label() string
	SetLabel(string)
	Filename() string
	IsEncrypted() bool
	SetEncrypted(cryptoType CryptoType, encryptedSecrets string)
	SetDecrypted()
	CryptoType() CryptoType
	Version() string
	SetVersion(string)
	AddressConstructor() func(cipher.PubKey) cipher.Addresser
	Secrets() string
	XPub() string

	UnpackSecrets(ss Secrets) error
	PackSecrets(ss Secrets)

	Erase()
	Clone() Wallet
	CopyFrom(src Wallet)
	CopyFromRef(src Wallet)

	ToReadable() Readable

	Validate() error

	Fingerprint() string
	GetAddresses() []cipher.Addresser
	GetSkycoinAddresses() ([]cipher.Address, error)
	GetEntryAt(i int) Entry
	GetEntry(cipher.Address) (Entry, bool)
	HasEntry(cipher.Address) bool
	EntriesLen() int
	GetEntries() Entries

	GenerateAddresses(num uint64) ([]cipher.Addresser, error)
	GenerateSkycoinAddresses(num uint64) ([]cipher.Address, error)
	ScanAddresses(scanN uint64, tf TransactionsFinder) error
}

// GuardUpdate executes a function within the context of a read-write managed decrypted wallet.
// Returns ErrWalletNotEncrypted if wallet is not encrypted.
func GuardUpdate(w Wallet, password []byte, fn func(w Wallet) error) error {
	if !w.IsEncrypted() {
		return ErrWalletNotEncrypted
	}

	if len(password) == 0 {
		return ErrMissingPassword
	}

	cryptoType := w.CryptoType()
	wlt, err := Unlock(w, password)
	if err != nil {
		return err
	}

	defer wlt.Erase()

	if err := fn(wlt); err != nil {
		return err
	}

	if err := Lock(wlt, password, cryptoType); err != nil {
		return err
	}

	w.CopyFromRef(wlt)

	// Wipes all sensitive data
	w.Erase()
	return nil
}

// GuardView executes a function within the context of a read-only managed decrypted wallet.
// Returns ErrWalletNotEncrypted if wallet is not encrypted.
func GuardView(w Wallet, password []byte, f func(w Wallet) error) error {
	if !w.IsEncrypted() {
		return ErrWalletNotEncrypted
	}

	if len(password) == 0 {
		return ErrMissingPassword
	}

	wlt, err := Unlock(w, password)
	if err != nil {
		return err
	}

	defer wlt.Erase()

	return f(wlt)
}

type walletLoadMeta struct {
	Meta struct {
		Type string `json:"type"`
	} `json:"meta"`
}

type walletLoader interface {
	SetFilename(string)
	SetCoin(CoinType)
	Coin() CoinType
	ToWallet() (Wallet, error)
}

// Load loads wallet from a given file
func Load(filename string) (Wallet, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("wallet %q doesn't exist", filename)
	}

	// Load the wallet meta type field from JSON
	var m walletLoadMeta
	if err := file.LoadJSON(filename, &m); err != nil {
		logger.WithError(err).WithField("filename", filename).Error("Load: file.LoadJSON failed")
		return nil, err
	}

	if !IsValidWalletType(m.Meta.Type) {
		logger.WithError(ErrInvalidWalletType).WithFields(logrus.Fields{
			"filename":   filename,
			"walletType": m.Meta.Type,
		}).Error("wallet meta loaded from disk has invalid wallet type")
		return nil, fmt.Errorf("invalid wallet %q: %v", filename, ErrInvalidWalletType)
	}

	// Depending on the wallet type in the wallet metadata header, load the full wallet data
	var rw walletLoader
	var err error
	switch m.Meta.Type {
	case WalletTypeDeterministic:
		logger.WithField("filename", filename).Info("LoadReadableDeterministicWallet")
		rw, err = LoadReadableDeterministicWallet(filename)
	case WalletTypeCollection:
		logger.WithField("filename", filename).Info("LoadReadableCollectionWallet")
		rw, err = LoadReadableCollectionWallet(filename)
	case WalletTypeBip44:
		logger.WithField("filename", filename).Info("LoadReadableBip44Wallet")
		rw, err = LoadReadableBip44Wallet(filename)
	case WalletTypeXPub:
		logger.WithField("filename", filename).Info("LoadReadableXPubWallet")
		rw, err = LoadReadableXPubWallet(filename)
	default:
		err := errors.New("unhandled wallet type")
		logger.WithField("walletType", m.Meta.Type).WithError(err).Error("Load failed")
		return nil, err
	}

	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"filename":   filename,
			"walletType": m.Meta.Type,
		}).Error("Load readable wallet failed")
		return nil, err
	}

	// Make sure "sky", "btc" normalize to "skycoin", "bitcoin"
	ct, err := ResolveCoinType(string(rw.Coin()))
	if err != nil {
		logger.WithError(err).WithField("coinType", rw.Coin()).Error("Load: invalid coin type")
		return nil, fmt.Errorf("invalid wallet %q: %v", filename, err)
	}
	rw.SetCoin(ct)

	rw.SetFilename(filepath.Base(filename))

	return rw.ToWallet()
}

// Save saves the wallet to a directory. The wallet's filename is read from its metadata.
func Save(w Wallet, dir string) error {
	rw := w.ToReadable()
	return file.SaveJSON(filepath.Join(dir, rw.Filename()), rw, 0600)
}

// removeBackupFiles removes any *.wlt.bak files whom have version 0.1 and *.wlt matched in the given directory
func removeBackupFiles(dir string) error {
	fs, err := filterDir(dir, ".wlt")
	if err != nil {
		return err
	}

	// Creates the .wlt file map
	fm := make(map[string]struct{})
	for _, f := range fs {
		fm[f] = struct{}{}
	}

	// Filters all .wlt.bak files in the directory
	bakFs, err := filterDir(dir, ".wlt.bak")
	if err != nil {
		return err
	}

	// Removes the .wlt.bak file that has .wlt matched.
	for _, bf := range bakFs {
		f := strings.TrimRight(bf, ".bak")
		if _, ok := fm[f]; ok {
			// Load and check the wallet version
			w, err := Load(f)
			if err != nil {
				return err
			}

			if w.Version() == "0.1" {
				if err := os.Remove(bf); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func filterDir(dir string, suffix string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), suffix) {
			res = append(res, filepath.Join(dir, f.Name()))
		}
	}
	return res, nil
}
