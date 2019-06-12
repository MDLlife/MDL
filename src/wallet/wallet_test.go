package wallet

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/MDLlife/MDL/src/cipher"
	"github.com/MDLlife/MDL/src/cipher/encrypt"
	"github.com/MDLlife/MDL/src/util/logging"
)

var (
	log = logging.MustGetLogger("wallet_test")
)

// set rand seed.
var _ = func() int64 {
	t := time.Now().Unix()
	rand.Seed(t)
	return t
}()

var u = flag.Bool("u", false, "update test wallet file in ./testdata")

func init() {
	flag.Parse()

	// Change the scrypt N value in cryptoTable to make test faster, otherwise
	// it would take more than 200 seconds to finish.
	cryptoTable[CryptoTypeScryptChacha20poly1305] = encrypt.ScryptChacha20poly1305{
		N:      1 << 15,
		R:      encrypt.ScryptR,
		P:      encrypt.ScryptP,
		KeyLen: encrypt.ScryptKeyLen,
	}

	// When -u flag is specified, update the following wallet files:
	//     - ./testdata/scrypt-chacha20poly1305-encrypted.wlt
	//     - ./testdata/sha256xor-encrypted.wlt
	if *u {
		// Update ./testdata/scrypt-chacha20poly1305-encrypted.wlt
		//     - Create an unencrypted wallet
		//     - Generate an address
		//     - Lock the wallet with scrypt-chacha20poly1305 crypto type and password of "pwd".
		w, err := NewWallet("scrypt-chacha20poly1305-encrypted.wlt", Options{
			Seed:  "seed-scrypt-chacha20poly1305",
			Label: "scrypt-chacha20poly1305",
		})
		if err != nil {
			log.Panic(err)
		}

		if _, err := w.GenerateAddresses(1); err != nil {
			log.Panic(err)
		}

		if err := w.Lock([]byte("pwd"), CryptoTypeScryptChacha20poly1305); err != nil {
			log.Panic(err)
		}

		if err := w.Save("./testdata"); err != nil {
			log.Panic(err)
		}

		// Update ./testdata/sha256xor-encrypted.wlt
		//     - Create an sha256xor encrypted wallet with password: "pwd".
		w1, err := NewWallet("sha256xor-encrypted.wlt", Options{
			Seed:       "seed-sha256xor",
			Label:      "sha256xor",
			Encrypt:    true,
			Password:   []byte("pwd"),
			CryptoType: CryptoTypeSha256Xor,
		})
		if err != nil {
			log.Panic(err)
		}

		if err := w1.Save("./testdata"); err != nil {
			log.Panic(err)
		}
	}
}

type mockBalanceGetter map[cipher.Address]BalancePair

func (mb mockBalanceGetter) GetBalanceOfAddrs(addrs []cipher.Address) ([]BalancePair, error) {
	var bals []BalancePair
	for _, addr := range addrs {
		bal := mb[addr]
		bals = append(bals, bal)
	}
	return bals, nil
}

func TestNewWallet(t *testing.T) {
	type expect struct {
		meta map[string]string
		err  error
	}

	tt := []struct {
		name    string
		wltName string
		ops     Options
		expect  expect
	}{
		{
			"ok with seed set",
			"test.wlt",
			Options{
				Seed: "testseed123",
			},
			expect{
				meta: map[string]string{
					"label":    "",
					"filename": "test.wlt",
					"coin":     string(CoinTypeMDL),
					"type":     WalletTypeDeterministic,
					"seed":     "testseed123",
					"version":  Version,
				},
				err: nil,
			},
		},
		{
			"ok with label and seed set",
			"test.wlt",
			Options{
				Label: "wallet1",
				Seed:  "testseed123",
			},
			expect{
				meta: map[string]string{
					"label":    "wallet1",
					"filename": "test.wlt",
					"coin":     string(CoinTypeMDL),
					"type":     WalletTypeDeterministic,
					"seed":     "testseed123",
					"version":  Version,
				},
				err: nil,
			},
		},
		{
			"ok with label, seed and coin set",
			"test.wlt",
			Options{
				Label: "wallet1",
				Coin:  CoinTypeBitcoin,
				Seed:  "testseed123",
			},
			expect{
				meta: map[string]string{
					"label":    "wallet1",
					"filename": "test.wlt",
					"coin":     string(CoinTypeBitcoin),
					"type":     WalletTypeDeterministic,
					"seed":     "testseed123",
				},
				err: nil,
			},
		},
		{
			"ok default crypto type",
			"test.wlt",
			Options{
				Label:    "wallet1",
				Coin:     CoinTypeMDL,
				Seed:     "testseed123",
				Encrypt:  true,
				Password: []byte("pwd"),
			},
			expect{
				meta: map[string]string{
					"label":     "wallet1",
					"coin":      string(CoinTypeMDL),
					"type":      WalletTypeDeterministic,
					"encrypted": "true",
				},
				err: nil,
			},
		},
		{
			"encrypt without password",
			"test.wlt",
			Options{
				Label:   "wallet1",
				Coin:    CoinTypeMDL,
				Seed:    "testseed123",
				Encrypt: true,
			},
			expect{
				meta: map[string]string{
					"label":     "wallet1",
					"coin":      string(CoinTypeMDL),
					"type":      WalletTypeDeterministic,
					"encrypted": "true",
				},
				err: ErrMissingPassword,
			},
		},
		{
			"create with no seed",
			"test.wlt",
			Options{
				Label:    "wallet1",
				Coin:     CoinTypeMDL,
				Encrypt:  true,
				Password: []byte("pwd"),
			},
			expect{
				meta: map[string]string{
					"label":     "wallet1",
					"coin":      string(CoinTypeMDL),
					"type":      WalletTypeDeterministic,
					"encrypted": "true",
				},
				err: ErrMissingSeed,
			},
		},
		{
			"password=pwd encrypt=false",
			"test.wlt",
			Options{
				Label:    "wallet1",
				Coin:     CoinTypeMDL,
				Encrypt:  false,
				Seed:     "seed",
				Password: []byte("pwd"),
			},
			expect{
				err: ErrMissingEncrypt,
			},
		},
	}

	for _, tc := range tt {
		// test all supported crypto types
		for ct := range cryptoTable {
			name := fmt.Sprintf("%v crypto=%v", tc.name, ct)
			if tc.ops.Encrypt {
				tc.ops.CryptoType = ct
			}
			t.Run(name, func(t *testing.T) {
				w, err := NewWallet(tc.wltName, tc.ops)
				require.Equal(t, tc.expect.err, err)
				if err != nil {
					return
				}

				require.Equal(t, tc.ops.Encrypt, w.IsEncrypted())

				if w.IsEncrypted() {
					// Confirms the seeds and entry secrets are all empty
					require.Equal(t, "", w.seed())
					require.Equal(t, "", w.lastSeed())

					for _, e := range w.Entries {
						require.True(t, e.Secret.Null())
					}

					// Confirms that secrets field is not empty
					require.NotEmpty(t, w.secrets())
				}
			})
		}
	}
}

func TestWalletLock(t *testing.T) {
	tt := []struct {
		name    string
		opts    Options
		lockPwd []byte
		err     error
	}{
		{
			"ok",
			Options{
				Seed: "seed",
			},
			[]byte("pwd"),
			nil,
		},
		{
			"password is nil",
			Options{
				Seed: "seed",
			},
			nil,
			ErrMissingPassword,
		},
		{
			"wallet already encrypted",
			Options{
				Seed:     "seed",
				Encrypt:  true,
				Password: []byte("pwd"),
			},
			[]byte("pwd"),
			ErrWalletEncrypted,
		},
	}

	for _, tc := range tt {
		for ct := range cryptoTable {
			name := fmt.Sprintf("%v crypto=%v", tc.name, ct)
			if tc.opts.Encrypt {
				tc.opts.CryptoType = ct
			}
			t.Run(name, func(t *testing.T) {
				wltName := NewWalletFilename()
				w, err := NewWallet(wltName, tc.opts)
				require.NoError(t, err)

				if !w.IsEncrypted() {
					// Generates 2 addresses
					_, err = w.GenerateAddresses(2)
					require.NoError(t, err)
				}

				err = w.Lock(tc.lockPwd, ct)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				require.True(t, w.IsEncrypted())

				// Checks if the seeds are wiped
				require.Empty(t, w.seed())
				require.Empty(t, w.lastSeed())

				// Checks if the entries are encrypted
				for i := range w.Entries {
					require.Equal(t, cipher.SecKey{}, w.Entries[i].Secret)
				}
			})

		}
	}

}

func TestWalletUnlock(t *testing.T) {
	tt := []struct {
		name      string
		opts      Options
		unlockPwd []byte
		err       error
	}{
		{
			"ok",
			Options{
				Seed:     "seed",
				Encrypt:  true,
				Password: []byte("pwd"),
			},
			[]byte("pwd"),
			nil,
		},
		{
			"unlock with nil password",
			Options{
				Seed:     "seed",
				Encrypt:  true,
				Password: []byte("pwd"),
			},
			nil,
			ErrMissingPassword,
		},
		{
			"unlock undecrypted wallet",
			Options{
				Seed:    "seed",
				Encrypt: false,
			},
			[]byte("pwd"),
			ErrWalletNotEncrypted,
		},
	}

	for _, tc := range tt {
		for ct := range cryptoTable {
			name := fmt.Sprintf("%v crypto=%v", tc.name, ct)
			if tc.opts.Encrypt {
				tc.opts.CryptoType = ct
			}
			t.Run(name, func(t *testing.T) {
				w := makeWallet(t, tc.opts, 1)
				// Tests the unlock method
				wlt, err := w.Unlock(tc.unlockPwd)
				require.Equal(t, tc.err, err)
				if err != nil {
					return
				}

				require.False(t, wlt.IsEncrypted())

				// Checks the seeds
				require.Equal(t, tc.opts.Seed, wlt.seed())

				// Checks the generated addresses
				sd, sks := cipher.MustGenerateDeterministicKeyPairsSeed([]byte(wlt.seed()), 1)
				require.Equal(t, uint64(1), uint64(len(wlt.Entries)))

				// Checks the last seed
				require.Equal(t, hex.EncodeToString(sd), wlt.lastSeed())

				for i := range wlt.Entries {
					addr := cipher.MustAddressFromSecKey(sks[i])
					require.Equal(t, addr, wlt.Entries[i].Address)
				}

				// Checks the original seeds
				require.NotEqual(t, tc.opts.Seed, w.seed())

				// Checks if the seckeys in entries of original wallet are empty
				for i := range w.Entries {
					require.Equal(t, cipher.SecKey{}, w.Entries[i].Secret)
				}

				// Checks if the seed and lastSeed in original wallet are sitll empty
				require.Empty(t, w.seed())
				require.Empty(t, w.lastSeed())
			})
		}
	}
}

func TestLockAndUnLock(t *testing.T) {
	for ct := range cryptoTable {
		t.Run(fmt.Sprintf("crypto=%v", ct), func(t *testing.T) {
			w, err := NewWallet("wallet", Options{
				Label: "wallet",
				Seed:  "seed",
			})
			require.NoError(t, err)
			_, err = w.GenerateAddresses(9)
			require.NoError(t, err)
			require.Len(t, w.Entries, 10)

			// clone the wallet
			cw := w.clone()
			require.Equal(t, w, cw)

			// lock the cloned wallet
			err = cw.Lock([]byte("pwd"), ct)
			require.NoError(t, err)

			// unlock the cloned wallet
			ucw, err := cw.Unlock([]byte("pwd"))
			require.NoError(t, err)

			require.Equal(t, w, ucw)
		})
	}
}

func makeWallet(t *testing.T, opts Options, addrNum uint64) *Wallet { // nolint: unparam
	// Create an unlocked wallet, then generate addresses, lock if the options.Encrypt is true.
	preOpts := opts
	opts.Encrypt = false
	opts.Password = nil
	w, err := NewWallet("t.wlt", opts)
	require.NoError(t, err)

	if addrNum > 1 {
		_, err = w.GenerateAddresses(addrNum - 1)
		require.NoError(t, err)
	}
	if preOpts.Encrypt {
		err = w.Lock(preOpts.Password, preOpts.CryptoType)
		require.NoError(t, err)
	}
	return w
}

func TestLoadWallet(t *testing.T) {
	type expect struct {
		meta map[string]string
		err  error
	}

	tt := []struct {
		name   string
		file   string
		expect expect
	}{
		{
			"ok",
			"./testdata/test1.wlt",
			expect{
				meta: map[string]string{
					"coin":     string(CoinTypeMDL),
					"filename": "test1.wlt",
					"label":    "test3",
					"lastSeed": "9182b02c0004217ba9a55593f8cf0abecc30d041e094b266dbb5103e1919adaf",
					"seed":     "buddy fossil side modify turtle door label grunt baby worth brush master",
					"tm":       "1503458909",
					"type":     WalletTypeDeterministic,
					"version":  "0.1",
				},
				err: nil,
			},
		},
		{
			"wallet file doesn't exist",
			"not_exist_file.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("wallet not_exist_file.wlt doesn't exist"),
			},
		},
		{
			"invalid wallet: no type",
			"./testdata/invalid_wallets/no_type.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("invalid wallet no_type.wlt: type field not set"),
			},
		},
		{
			"invalid wallet: invalid type",
			"./testdata/invalid_wallets/err_type.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("invalid wallet err_type.wlt: wallet type invalid"),
			},
		},
		{
			"invalid wallet: no coin",
			"./testdata/invalid_wallets/no_coin.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("invalid wallet no_coin.wlt: coin field not set"),
			},
		},
		{
			"invalid wallet: no seed",
			"./testdata/invalid_wallets/no_seed.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("invalid wallet no_seed.wlt: seed missing in unencrypted wallet"),
			},
		},
		{
			"version=0.2 encrypted=true crypto=scrypt-chacha20poly1305",
			"./testdata/scrypt-chacha20poly1305-encrypted.wlt",
			expect{
				meta: map[string]string{
					"coin":       string(CoinTypeMDL),
					"cryptoType": "scrypt-chacha20poly1305",
					"encrypted":  "true",
					"filename":   "scrypt-chacha20poly1305-encrypted.wlt",
					"label":      "scrypt-chacha20poly1305",
					"lastSeed":   "",
					"seed":       "",
					"type":       WalletTypeDeterministic,
					"version":    "0.2",
				},
				err: nil,
			},
		},
		{
			"version=0.2 encrypted=true crypto=sha256xor",
			"./testdata/sha256xor-encrypted.wlt",
			expect{
				meta: map[string]string{
					"coin":       string(CoinTypeMDL),
					"cryptoType": "sha256-xor",
					"encrypted":  "true",
					"filename":   "sha256xor-encrypted.wlt",
					"label":      "sha256xor",
					"lastSeed":   "",
					"seed":       "",
					"type":       WalletTypeDeterministic,
					"version":    "0.2",
				},
				err: nil,
			},
		},
		{
			"version=0.2 encrypted=false",
			"./testdata/v2_no_encrypt.wlt",
			expect{
				meta: map[string]string{
					"coin":       string(CoinTypeMDL),
					"cryptoType": "scrypt-chacha20poly1305",
					"encrypted":  "false",
					"filename":   "v2_no_encrypt.wlt",
					"label":      "v2_no_encrypt",
					"lastSeed":   "c79454cf362b3f55e5effce09f664311650a44b9c189b3c8eed1ae9bd696cd9e",
					"secrets":    "",
					"seed":       "seed",
					"type":       WalletTypeDeterministic,
					"version":    "0.2",
				},
				err: nil,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w, err := Load(tc.file)
			require.Equal(t, tc.expect.err, err)
			if err != nil {
				return
			}

			for k, v := range tc.expect.meta {
				vv := w.Meta[k]
				require.Equal(t, v, vv)
			}

			if w.IsEncrypted() {
				require.NotEmpty(t, w.Meta[metaSecrets])
			}
		})
	}
}

func TestWalletGenerateAddress(t *testing.T) {
	tt := []struct {
		name               string
		opts               Options
		num                uint64
		oneAddressEachTime bool
		err                error
	}{
		{
			"ok with one address",
			Options{
				Seed: "seed",
			},
			1,
			false,
			nil,
		},
		{
			"ok with two address",
			Options{
				Seed: "seed",
			},
			2,
			false,
			nil,
		},
		{
			"ok with three address and generate one address each time",
			Options{
				Seed: "seed",
			},
			2,
			true,
			nil,
		},
		{
			"wallet is encrypted",
			Options{
				Seed:     "seed",
				Encrypt:  true,
				Password: []byte("pwd"),
			},
			2,
			true,
			ErrWalletEncrypted,
		},
	}

	for _, tc := range tt {
		for ct := range cryptoTable {
			name := fmt.Sprintf("crypto=%v %v", ct, tc.name)
			if tc.opts.Encrypt {
				tc.opts.CryptoType = ct
			}

			t.Run(name, func(t *testing.T) {
				// create wallet
				w, err := NewWallet("test.wlt", tc.opts)
				require.NoError(t, err)

				// generate addresses
				if tc.oneAddressEachTime {
					_, err = w.GenerateAddresses(tc.num - 1)
					require.Equal(t, tc.err, err)
					if err != nil {
						return
					}
				} else {
					for i := uint64(0); i < tc.num-1; i++ {
						_, err := w.GenerateAddresses(1)
						require.Equal(t, tc.err, err)
						if err != nil {
							return
						}
					}
				}

				// check the entry number
				require.Equal(t, int(tc.num), len(w.Entries))

				addrs := w.GetAddresses()

				_, keys := cipher.MustGenerateDeterministicKeyPairsSeed([]byte(tc.opts.Seed), int(tc.num))
				for i, k := range keys {
					a := cipher.MustAddressFromSecKey(k)
					require.Equal(t, a.String(), addrs[i].String())
				}
			})
		}
	}
}

func TestWalletGetEntry(t *testing.T) {
	tt := []struct {
		name    string
		wltFile string
		address string
		find    bool
	}{
		{
			"ok",
			"./testdata/test1.wlt",
			"JUdRuTiqD1mGcw358twMg3VPpXpzbkdRvJ",
			true,
		},
		{
			"entry not exist",
			"./testdata/test1.wlt",
			"2ULfxDUuenUY5V4Pr8whmoAwFdUseXNyjXC",
			false,
		},
		{
			"scrypt-chacha20poly1305 encrytped wallet",
			"./testdata/scrypt-chacha20poly1305-encrypted.wlt",
			"LxcitUpWQZbPjgEPs6R1i3G4Xa31nPMoSG",
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w, err := Load(tc.wltFile)
			require.NoError(t, err)
			a, err := cipher.DecodeBase58Address(tc.address)
			require.NoError(t, err)
			e, ok := w.GetEntry(a)
			require.Equal(t, tc.find, ok)
			if ok {
				require.Equal(t, tc.address, e.Address.String())
			}
		})
	}
}

func TestWalletAddEntry(t *testing.T) {
	test1SecKey, err := cipher.SecKeyFromHex("1fc5396e91e60b9fc613d004ea5bd2ccea17053a12127301b3857ead76fdb93e")
	require.NoError(t, err)

	_, s := cipher.GenerateKeyPair()
	seckeys := []cipher.SecKey{
		test1SecKey,
		s,
	}

	tt := []struct {
		name    string
		wltFile string
		secKey  cipher.SecKey
		err     error
	}{
		{
			"ok",
			"./testdata/test1.wlt",
			seckeys[1],
			nil,
		},
		{
			"dup entry",
			"./testdata/test1.wlt",
			seckeys[0],
			errors.New("duplicate address entry"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w, err := Load(tc.wltFile)
			require.NoError(t, err)
			a := cipher.MustAddressFromSecKey(tc.secKey)
			p := cipher.MustPubKeyFromSecKey(tc.secKey)
			require.Equal(t, tc.err, w.AddEntry(Entry{
				Address: a,
				Public:  p,
				Secret:  s,
			}))
		})
	}
}

func TestWalletGuard(t *testing.T) {
	for ct := range cryptoTable {
		t.Run(fmt.Sprintf("crypto=%v", ct), func(t *testing.T) {
			validate := func(w *Wallet) {
				require.Equal(t, "", w.seed())
				require.Equal(t, "", w.lastSeed())
				for _, e := range w.Entries {
					require.Equal(t, cipher.SecKey{}, e.Secret)
				}
			}

			w, err := NewWallet("t.wlt", Options{
				Seed:       "seed",
				Encrypt:    true,
				Password:   []byte("pwd"),
				CryptoType: ct,
			})
			require.NoError(t, err)

			err = w.GuardUpdate([]byte("pwd"), func(w *Wallet) error {
				require.Equal(t, "seed", w.seed())
				w.setLabel("label")
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, "label", w.Label())
			validate(w)

			err = w.GuardView([]byte("pwd"), func(w *Wallet) error {
				require.Equal(t, "label", w.Label())
				w.setLabel("new label")
				return nil
			})
			require.NoError(t, err)

			require.Equal(t, "label", w.Label())
			validate(w)

		})
	}
}

func TestRemoveBackupFiles(t *testing.T) {
	type wltInfo struct {
		wltName string
		version string
	}

	tt := []struct {
		name                   string
		initFiles              []wltInfo
		expectedRemainingFiles map[string]struct{}
	}{
		{
			"no file",
			[]wltInfo{},
			map[string]struct{}{},
		},
		{
			"wlt v0.1=1 bak v0.1=1 delete 1 bak",
			[]wltInfo{
				{
					"t1.wlt",
					"0.1",
				},
				{
					"t1.wlt.bak",
					"0.1",
				},
			},
			map[string]struct{}{
				"t1.wlt": {},
			},
		},
		{
			"wlt v0.1=2 bak v0.1=1 delete 1 bak",
			[]wltInfo{
				{
					"t1.wlt",
					"0.1",
				},
				{
					"t2.wlt",
					"0.1",
				},
				{
					"t2.wlt.bak",
					"0.1",
				},
			},
			map[string]struct{}{
				"t1.wlt": {},
				"t2.wlt": {},
			},
		},
		{
			"wlt v0.1=3 bak v0.1=1 delete 1 bak",
			[]wltInfo{
				{
					"t1.wlt",
					"0.1",
				},
				{
					"t2.wlt",
					"0.1",
				},
				{
					"t3.wlt",
					"0.1",
				},
				{
					"t3.wlt.bak",
					"0.1",
				},
			},
			map[string]struct{}{
				"t1.wlt": {},
				"t2.wlt": {},
				"t3.wlt": {},
			},
		},
		{
			"wlt v0.1=3 bak v0.1=2 delete 2 bak",
			[]wltInfo{
				{
					"t1.wlt",
					"0.1",
				},
				{
					"t2.wlt",
					"0.1",
				},
				{
					"t2.wlt.bak",
					"0.1",
				},
				{
					"t3.wlt",
					"0.1",
				},
				{
					"t3.wlt.bak",
					"0.1",
				},
			},
			map[string]struct{}{
				"t1.wlt": {},
				"t2.wlt": {},
				"t3.wlt": {},
			},
		},
		{
			"wlt v0.1=3 bak v0.1=3 delete 3 bak",
			[]wltInfo{
				{
					"t1.wlt",
					"0.1",
				},
				{
					"t1.wlt.bak",
					"0.1",
				},
				{
					"t2.wlt",
					"0.1",
				},
				{
					"t2.wlt.bak",
					"0.1",
				},
				{
					"t3.wlt",
					"0.1",
				},
				{
					"t3.wlt.bak",
					"0.1",
				},
			},
			map[string]struct{}{
				"t1.wlt": {},
				"t2.wlt": {},
				"t3.wlt": {},
			},
		},
		{
			"wlt v0.1=3 bak v0.1=1 no delete",
			[]wltInfo{
				{
					"t1.wlt",
					"0.1",
				},
				{
					"t2.wlt",
					"0.1",
				},
				{
					"t3.wlt",
					"0.1",
				},
				{
					"t4.wlt.bak",
					"0.1",
				},
			},
			map[string]struct{}{
				"t1.wlt":     {},
				"t2.wlt":     {},
				"t3.wlt":     {},
				"t4.wlt.bak": {},
			},
		},
		{
			"wlt v0.2=3 bak v0.2=1 no delete",
			[]wltInfo{
				{
					"t1.wlt",
					"0.2",
				},
				{
					"t2.wlt",
					"0.2",
				},
				{
					"t3.wlt",
					"0.2",
				},
				{
					"t3.wlt.bak",
					"0.2",
				},
			},
			map[string]struct{}{
				"t1.wlt":     {},
				"t2.wlt":     {},
				"t3.wlt":     {},
				"t3.wlt.bak": {},
			},
		},
		{
			"wlt v0.1=1 bak v0.1=1 wlt v0.2=2 bak v0.2=2 delete 1 bak",
			[]wltInfo{
				{
					"t1.wlt",
					"0.1",
				},
				{
					"t1.wlt.bak",
					"0.1",
				},
				{
					"t2.wlt",
					"0.2",
				},
				{
					"t2.wlt.bak",
					"0.2",
				},
				{
					"t3.wlt",
					"0.2",
				},
				{
					"t3.wlt.bak",
					"0.2",
				},
			},
			map[string]struct{}{
				"t1.wlt":     {},
				"t2.wlt":     {},
				"t2.wlt.bak": {},
				"t3.wlt":     {},
				"t3.wlt.bak": {},
			},
		},
		{
			"wlt v0.1=1 bak v0.1=2 wlt v0.2=2 bak v0.2=1 delete 1 bak",
			[]wltInfo{
				{
					"t1.wlt",
					"0.1",
				},
				{
					"t1.wlt.bak",
					"0.1",
				},
				{
					"t2.wlt",
					"0.2",
				},
				{
					"t2.wlt.bak",
					"0.1",
				},
				{
					"t3.wlt",
					"0.2",
				},
				{
					"t3.wlt.bak",
					"0.2",
				},
			},
			map[string]struct{}{
				"t1.wlt":     {},
				"t2.wlt":     {},
				"t2.wlt.bak": {},
				"t3.wlt":     {},
				"t3.wlt.bak": {},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dir := prepareWltDir()
			// Initialize files
			for _, f := range tc.initFiles {
				w, err := NewWallet(f.wltName, Options{
					Seed: "s1",
				})
				require.NoError(t, err)
				w.setVersion(f.version)

				require.NoError(t, w.Save(dir))
			}

			require.NoError(t, removeBackupFiles(dir))

			// Get all remaining files
			fs, err := ioutil.ReadDir(dir)
			require.NoError(t, err)
			require.Len(t, fs, len(tc.expectedRemainingFiles))
			for _, f := range fs {
				_, ok := tc.expectedRemainingFiles[f.Name()]
				require.True(t, ok)
			}
		})
	}
}

func TestWalletValidate(t *testing.T) {
	goodMetaUnencrypted := map[string]string{
		"filename":  "foo.wlt",
		"type":      WalletTypeDeterministic,
		"coin":      string(CoinTypeMDL),
		"encrypted": "false",
		"seed":      "fooseed",
		"lastSeed":  "foolastseed",
	}

	goodMetaEncrypted := map[string]string{
		"filename":   "foo.wlt",
		"type":       WalletTypeDeterministic,
		"coin":       string(CoinTypeMDL),
		"encrypted":  "true",
		"cryptoType": "scrypt-chacha20poly1305",
		"seed":       "",
		"lastSeed":   "",
		"secrets":    "xacsdasdasdasd",
	}

	copyMap := func(m map[string]string) map[string]string {
		n := make(map[string]string, len(m))
		for k, v := range m {
			n[k] = v
		}
		return n
	}

	delField := func(m map[string]string, f string) map[string]string {
		n := copyMap(m)
		delete(n, f)
		return n
	}

	setField := func(m map[string]string, f, g string) map[string]string {
		n := copyMap(m)
		n[f] = g
		return n
	}

	cases := []struct {
		name string
		meta map[string]string
		err  error
	}{
		{
			name: "missing filename",
			meta: delField(goodMetaUnencrypted, metaFilename),
			err:  errors.New("filename not set"),
		},
		{
			name: "wallet type missing",
			meta: delField(goodMetaUnencrypted, metaType),
			err:  errors.New("type field not set"),
		},
		{
			name: "wallet type invalid",
			meta: setField(goodMetaUnencrypted, metaType, "footype"),
			err:  errors.New("wallet type invalid"),
		},
		{
			name: "coin field missing",
			meta: delField(goodMetaUnencrypted, metaCoin),
			err:  errors.New("coin field not set"),
		},
		{
			name: "encrypted field invalid",
			meta: setField(goodMetaUnencrypted, metaEncrypted, "foo"),
			err:  errors.New("encrypted field is not a valid bool"),
		},
		{
			name: "unencrypted missing seed",
			meta: delField(goodMetaUnencrypted, metaSeed),
			err:  errors.New("seed missing in unencrypted wallet"),
		},
		{
			name: "unencrypted missing last seed",
			meta: delField(goodMetaUnencrypted, metaLastSeed),
			err:  errors.New("lastSeed missing in unencrypted wallet"),
		},
		{
			name: "crypto type missing",
			meta: delField(goodMetaEncrypted, metaCryptoType),
			err:  errors.New("crypto type field not set"),
		},
		{
			name: "crypto type invalid",
			meta: setField(goodMetaEncrypted, metaCryptoType, "foocryptotype"),
			err:  errors.New("unknown crypto type"),
		},
		{
			name: "secrets missing",
			meta: delField(goodMetaEncrypted, metaSecrets),
			err:  errors.New("wallet is encrypted, but secrets field not set"),
		},
		{
			name: "secrets empty",
			meta: setField(goodMetaEncrypted, metaSecrets, ""),
			err:  errors.New("wallet is encrypted, but secrets field not set"),
		},
		{
			name: "valid unencrypted",
			meta: goodMetaUnencrypted,
		},
		{
			name: "valid encrypted",
			meta: goodMetaEncrypted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := &Wallet{
				Meta: tc.meta,
			}
			err := w.Validate()
			require.Equal(t, tc.err, err)
		})
	}
}
