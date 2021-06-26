package secp256k1

import (
	"bytes"
	"encoding/hex"
	"log"
	"testing"
)

const TESTS = 10000 // number of tests to use for random tests

func TestRandByte(t *testing.T) {
	nonce := RandByte(32)
	if len(nonce) != 32 {
		t.Fatal()
	}
}

func TestBitTwiddle(t *testing.T) {
	// test agreement for highest bit test
	var b byte
	for i := 0; i < 512; i++ {
		bool1 := ((b >> 7) == 1)
		bool2 := ((b & 0x80) == 0x80)
		if bool1 != bool2 {
			t.Fatal()
		}
		b++
	}
}

// assertSigCompact tests for signature malleability
// highest bit of S must be 0; 32nd byte
func assertSigCompact(sig []byte) {
	b := int(sig[32])
	if b < 0 {
		log.Panic()
	}
	if ((b >> 7) == 1) != ((b & 0x80) == 0x80) {
		log.Panicf("b=%v b2=%v", b, b>>7)
	}
	if (b & 0x80) == 0x80 {
		log.Panicf("b=%v b2=%v", b, b&0x80)
	}
}

func TestGenerateKeyPairValidKeys(t *testing.T) {
	pubkey, seckey := GenerateKeyPair()
	if VerifySeckey(seckey) != 1 {
		t.Fatal()
	}
	if VerifyPubkey(pubkey) != 1 {
		t.Fatal()
	}
}

func TestPubkeyFromSeckey(t *testing.T) {
	// test compressed pubkey from private key
	// http://www.righto.com/2014/02/bitcoins-hard-way-using-raw-bitcoin.html
	privkey, err := hex.DecodeString(`f19c523315891e6e15ae0608a35eec2e00ebd6d1984cf167f46336dabd9b2de4`)
	if err != nil {
		t.Fatal()
	}

	desiredPubKey, err := hex.DecodeString(`03fe43d0c2c3daab30f9472beb5b767be020b81c7cc940ed7a7e910f0c1d9feef1`)
	if err != nil {
		t.Fatal()
	}

	if pubkey := PubkeyFromSeckey(privkey); pubkey == nil {
		t.Fatal()
	} else if !bytes.Equal(pubkey, desiredPubKey) {
		t.Fatal()
	}
}

func TestUncompressedPubkeyFromSeckey(t *testing.T) {
	// test uncompressed pubkey from private key
	// http://www.righto.com/2014/02/bitcoins-hard-way-using-raw-bitcoin.html
	privkey, err := hex.DecodeString(`f19c523315891e6e15ae0608a35eec2e00ebd6d1984cf167f46336dabd9b2de4`)
	if err != nil {
		t.Fatal()
	}

	desiredPubKey, err := hex.DecodeString(`04fe43d0c2c3daab30f9472beb5b767be020b81c7cc940ed7a7e910f0c1d9feef10fe85eb3ce193405c2dd8453b7aeb6c1752361efdbf4f52ea8bf8f304aab37ab`)
	if err != nil {
		t.Fatal()
	}

	if pubkey := UncompressedPubkeyFromSeckey(privkey); pubkey == nil {
		t.Fatal()
	} else if !bytes.Equal(pubkey, desiredPubKey) {
		t.Fatal()
	}
}

// returns random pubkey, seckey, hash and signature
func randX() ([]byte, []byte, []byte, []byte) {
	pubkey, seckey := GenerateKeyPair()
	msg := RandByte(32)
	sig := Sign(msg, seckey)
	return pubkey, seckey, msg, sig
}

func TestSignatureVerifyPubkey(t *testing.T) {
	pubkey1, seckey := GenerateKeyPair()
	msg := RandByte(32)
	sig := Sign(msg, seckey)
	if VerifyPubkey(pubkey1) == 0 {
		t.Fail()
	}
	pubkey2 := RecoverPubkey(msg, sig)
	if !bytes.Equal(pubkey1, pubkey2) {
		t.Fatal("Recovered pubkey does not match")
	}
}

func TestVerifyFunctions(t *testing.T) {
	pubkey, seckey, hash, sig := randX()
	if VerifySeckey(seckey) == 0 {
		t.Fail()
	}
	if VerifyPubkey(pubkey) == 0 {
		t.Fail()
	}
	if VerifySignature(hash, sig, pubkey) == 0 {
		t.Fail()
	}
}

func TestSignOutputSizes(t *testing.T) {
	// test size of messages
	pubkey, seckey := GenerateKeyPair()
	msg := RandByte(32)
	sig := Sign(msg, seckey)
	assertSigCompact(sig)
	if sig == nil {
		t.Fatal("Signature nil")
	}
	if len(pubkey) != 33 {
		t.Fail()
	}
	if len(seckey) != 32 {
		t.Fail()
	}
	if len(sig) != 64+1 {
		t.Fail()
	}
	if int(sig[64]) > 4 {
		t.Fail()
	} // recovery ID, should be 0 to 4
}

func TestRecoverVerifySignature1(t *testing.T) {
	// test signing message
	pubkey1, seckey := GenerateKeyPair()
	msg := RandByte(32)
	sig := Sign(msg, seckey)
	if sig == nil {
		t.Fatal("Signature nil")
	}

	pubkey2 := RecoverPubkey(msg, sig)
	if pubkey2 == nil {
		t.Fatal("Recovered pubkey invalid")
	}
	if !bytes.Equal(pubkey1, pubkey2) {
		t.Fatal("Recovered pubkey does not match")
	}

	ret := VerifySignature(msg, sig, pubkey1)
	if ret != 1 {
		t.Fatal("Signature invalid")
	}
}

func TestVerifySignatureFailure(t *testing.T) {
	p, s := GenerateKeyPair()
	msg := RandByte(32)
	sig := Sign(msg, s)
	if sig == nil {
		t.Fatal("Signature nil")
	}

	p2, _ := GenerateKeyPair()

	if bytes.Equal(p, p2) {
		t.Fatal("GenerateKeyPair returned the same keys twice")
	}

	// Signature has different pubkey
	ret := VerifySignature(msg, sig, p2)
	if ret == 1 {
		t.Fatal("Signature unexpectedly valid")
	}

	msg2 := RandByte(32)
	sig2 := Sign(msg2, s)

	// Signature is for a different message
	ret = VerifySignature(msg, sig2, p)
	if ret == 1 {
		t.Fatal("Signature unexpectedly valid")
	}

	// Signature is for a different message
	ret = VerifySignature(msg2, sig, p)
	if ret == 1 {
		t.Fatal("Signature unexpectedly valid")
	}

	// Signature is for a different message
	msg3 := RandByte(32)
	ret = VerifySignature(msg3, sig, p)
	if ret == 1 {
		t.Fatal("Signature unexpectedly valid")
	}
}

func TestRecoverVerifySignature2(t *testing.T) {
	// test pubkey recovery
	pubkey1, seckey1 := GenerateKeyPair()
	msg := RandByte(32)
	sig := Sign(msg, seckey1)

	if sig == nil {
		t.Fatal("Signature nil")
	}
	ret := VerifySignature(msg, sig, pubkey1)
	if ret != 1 {
		t.Fatal("Signature invalid")
	}

	pubkey2 := RecoverPubkey(msg, sig)
	if len(pubkey1) != len(pubkey2) {
		t.Fatal()
	}
	for i := range pubkey1 {
		if pubkey1[i] != pubkey2[i] {
			t.Fatal()
		}
	}
	if !bytes.Equal(pubkey1, pubkey2) {
		t.Fatal()
	}
}

func TestRecoverPubkey1(t *testing.T) {
	// test random messages for the same pub/private key
	_, seckey := GenerateKeyPair()
	for i := 0; i < TESTS; i++ {
		msg := RandByte(32)
		sig := Sign(msg, seckey)
		assertSigCompact(sig)

		sig[len(sig)-1] %= 4
		pubkey2 := RecoverPubkey(msg, sig)
		if pubkey2 == nil {
			t.Fail()
		}
	}
}

func TestRecoverPubkey2(t *testing.T) {
	// test random messages for different pub/private keys
	for i := 0; i < TESTS; i++ {
		pubkey1, seckey := GenerateKeyPair()
		msg := RandByte(32)
		sig := Sign(msg, seckey)
		assertSigCompact(sig)

		if sig[len(sig)-1] >= 4 {
			t.Fail()
		}
		pubkey2 := RecoverPubkey(msg, sig)
		if pubkey2 == nil {
			t.Fail()
		}
		if !bytes.Equal(pubkey1, pubkey2) {
			t.Fail()
		}
	}
}

func randSig() []byte {
	sig := RandByte(65)
	sig[32] &= 0x70
	sig[64] %= 4
	return sig
}

func TestRecoverVerifySignature3(t *testing.T) {
	pubkey1, seckey := GenerateKeyPair()
	msg := RandByte(32)
	sig := Sign(msg, seckey)

	if sig == nil {
		t.Fail()
	}
	if len(sig) != 65 {
		t.Fail()
	}
	for i := 0; i < TESTS; i++ {
		sig = randSig()
		pubkey2 := RecoverPubkey(msg, sig)

		if bytes.Equal(pubkey1, pubkey2) {
			t.Fail()
		}

		if pubkey2 != nil && VerifySignature(msg, sig, pubkey2) != 1 {
			t.Fail()
		}

		if VerifySignature(msg, sig, pubkey1) == 1 {
			t.Fail()
		}
	}
}

func TestRecoverVerifySignature4(t *testing.T) {
	// test random messages against valid signature: should fail
	pubkey1, seckey := GenerateKeyPair()
	msg := RandByte(32)
	sig := Sign(msg, seckey)

	failCount := 0
	for i := 0; i < TESTS; i++ {
		msg = RandByte(32)
		pubkey2 := RecoverPubkey(msg, sig)
		if bytes.Equal(pubkey1, pubkey2) {
			t.Fail()
		}

		if pubkey2 != nil && VerifySignature(msg, sig, pubkey2) != 1 {
			t.Fail()
		}

		if VerifySignature(msg, sig, pubkey1) == 1 {
			t.Fail()
		}
	}
	if failCount != 0 {
		t.Errorf("ERROR: Accepted signature for %v of %v random messages", failCount, TESTS)
	}
}

/*
	Deterministic Keypair Tests
*/

func TestDeterministicKeypairs00(t *testing.T) {
	for i := 0; i < 64; i++ {
		seed := RandByte(64)
		_, pub1, sec1 := DeterministicKeyPairIterator(seed)
		pub2, sec2 := GenerateDeterministicKeyPair(seed)

		if !bytes.Equal(pub1, pub2) {
			t.Fail()
		}
		if !bytes.Equal(sec1, sec2) {
			t.Fail()
		}
	}
}

func TestDeterministicKeypairs01(t *testing.T) {
	for i := 0; i < 64; i++ {
		seed := RandByte(32)
		_, pub1, sec1 := DeterministicKeyPairIterator(seed)
		pub2, sec2 := GenerateDeterministicKeyPair(seed)

		if !bytes.Equal(pub1, pub2) {
			t.Fail()
		}
		if !bytes.Equal(sec1, sec2) {
			t.Fail()
		}
	}
}

func TestDeterministicKeypairs02(t *testing.T) {
	for i := 0; i < 64; i++ {
		seed := RandByte(32)
		_, pub1, sec1 := DeterministicKeyPairIterator(seed)
		pub2, sec2 := GenerateDeterministicKeyPair(seed)

		if !bytes.Equal(pub1, pub2) {
			t.Fail()
		}
		if !bytes.Equal(sec1, sec2) {
			t.Fail()
		}
	}
}

func MustDecodeHex(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}

func TestDeterministicKeypairs03(t *testing.T) {
	cases := []struct {
		seed   string
		seckey string
		pubkey string
	}{
		{
			seed:   "tQ93w5Aqcunm9SGUfnmF4fJv",
			seckey: "9b8c3e36adce64dedc80d6dfe51ff1742cc1d755bbad457ac01177c5a18a789f",
			pubkey: "03996e65d79e957ce1eafb57453e55b55906e04c8de556e54961eb06a4836c06df",
		},
		{
			seed:   "DC7qdQQtbWSSaekXnFmvQgse",
			seckey: "d2deaf4a9ff7a5111fe1d429d6976cbde78811fdd075371a2a4449bb0f4d8bf9",
			pubkey: "0283a86efb1b8d82147c336c83d991f8124f0c4ca62c1019d6af1db46ae34594be",
		},
		{
			seed:   "X8EkuUZC7Td7PAXeS7Duc7vR",
			seckey: "cad79b6dcf7bd21891cbe20a51c57d59689ae6e3dc482cd6ec22898ac00cd86b",
			pubkey: "03f1fbd857b8a19de3ae35d616d41f179c0f3de94231e3caabf34eabf4674a1643",
		},
		{
			seed:   "tVqPYHHNVPRWyEed62v7f23u",
			seckey: "2a386e94e9ffaa409517cbed81b9b2d4e1c5fb4afe3cbd67ce8aba11af0b02fa",
			pubkey: "03ebde2c29e3beadab6f324ceb82a71c23655678e47d97f1d92159c3d7e4b59be4",
		},
		{
			seed:   "kCy4R57HDfLqF3pVhBWxuMcg",
			seckey: "26a7c6d8809c476a56f7455209f58b5ff3f16435fcf208ff2931ece60067f305",
			pubkey: "03b27bd3ae6b9034a4ffb2173381448c724f649fd0ec14ee0288758aa7812a7338",
		},
		{
			seed:   "j8bjv86ZNjKqzafR6mtSUVCE",
			seckey: "ea5c0f8c9f091a70bf38327adb9b2428a9293e7a7a75119920d759ecfa03a995",
			pubkey: "0236b5d52711f8a11da664c57da4378690751016ecf3089eed4ed1833c610046b6",
		},
		{
			seed:   "qShryAzVY8EtsuD3dsAc7qnG",
			seckey: "331206176509bcae31c881dc51e90a4e82ec33cd7208a5fb4171ed56602017fa",
			pubkey: "02da2aab09ec94e8a40d7381f72ff6585bf7d87e4a599d1408d2686ce5514692b1",
		},
		{
			seed:   "5FGG7ZBa8wVMBJkmzpXj5ESX",
			seckey: "4ea2ad82e7730d30c0c21d01a328485a0cf5543e095139ba613929be7739b52c",
			pubkey: "02b7d159de0d705c99e24d609b1591b1fac86d46c2c99e2ce6cc20b7402e32215c",
		},
		{
			seed:   "f46TZG4xJHXUGWx8ekbNqa9F",
			seckey: "dcddd403d3534c4ef5703cc07a771c107ed49b7e0643c6a2985a96149db26108",
			pubkey: "03e709fbdaf4f3b913b8c4ea887d1fea61ed356fcf0178ee7c2b556ce308cfc001",
		},
		{
			seed:   "XkZdQJ5LT96wshN8JBH8rvEt",
			seckey: "3e276219081f072dff5400ca29a9346421eaaf3c419ff1474ac1c81ad8a9d6e1",
			pubkey: "03a34782043386f068780cc82d0deffcea1703e4e4bbe67537a89bda0fbd3240e0",
		},
		{
			seed:   "GFDqXU4zYymhJJ9UGqRgS8ty",
			seckey: "95be4163085b571e725edeffa83fff8e7a7db3c1ccab19d0f3c6e105859b5e10",
			pubkey: "03ad4b4525c6031c2fa3c43722ca6dbde64b30d646b8914b0391096d8964e5d4da",
		},
		{
			seed:   "tmwZksH2XyvuamnddYxyJ5Lp",
			seckey: "2666dd54e469df56c02e82dffb4d3ea067daafe72c54dc2b4f08c4fb3a7b7e42",
			pubkey: "03f127118872ac5cb83b9167e561a28d82f4691d06297051dc71fb97d00b42aa20",
		},
		{
			seed:   "EuqZFsbAV5amTzkhgAMgjr7W",
			seckey: "40c325c01f2e4087fcc97fcdbea6c35c88a12259ebf1bce0b14a4d77f075abbf",
			pubkey: "03df10131650e63894e6c43427fc9ad0119a956ef1821c68f0e09e7d90411e6c39",
		},
		{
			seed:   "TW6j8rMffZfmhyDEt2JUCrLB",
			seckey: "e676e0685c5d1afd43ad823b83db5c6100135c35485146276ee0b0004bd6689e",
			pubkey: "03f9ead3620dfcfcf731d42b0752a2e1549b0190070eed686002e02f58da955731",
		},
		{
			seed:   "8rvkBnygfhWP8kjX9aXq68CY",
			seckey: "21450a646eed0d4aa50a1736e6c9bf99fff006a470aab813a2eff3ee4d460ae4",
			pubkey: "026ace328af3200b4abe13a29125545bd9d82cc32eed13b782371ef54fb6301d6c",
		},
		{
			seed:   "phyRfPDuf9JMRFaWdGh7NXPX",
			seckey: "ca7bc04196c504d0e815e125f7f1e086c8ae8c10d5e9df984aeab4b41bf9e398",
			pubkey: "03fc05f68ef56235b777168c45d46dfb8f665d12e4f92265305b2e66aafe000351",
		},
	}

	for _, tc := range cases {
		t.Run(tc.seed, func(t *testing.T) {
			seed := []byte(tc.seed)
			sec1 := MustDecodeHex(tc.seckey)
			pub1 := MustDecodeHex(tc.pubkey)

			pub2, sec2 := GenerateDeterministicKeyPair(seed)
			if !bytes.Equal(sec1, sec2) {
				t.Error("seckeys don't match")
			}

			if !bytes.Equal(pub1, pub2) {
				t.Errorf("pubkeys don't match: %s != %s", tc.pubkey, hex.EncodeToString(pub2))
			}
		})
	}
}

func TestDeterministicKeyPairIterator1(t *testing.T) {
	cases := []struct {
		seed     string
		seckey   string
		pubkey   string
		nextSeed string
	}{
		{
			seed:     "90c56f5b8d78a46fb4cddf6fd9c6d88d6d2d7b0ec35917c7dac12c03b04e444e",
			seckey:   "94dd1a9de9ffd57b5516b8a7f090da67f142f7d22356fa5d1b894ee4d4fba95b",
			pubkey:   "03eb71274ba23438f4ce6ac125e20bb78cd8123dc9483b5f34ace67cb6972e4ca8",
			nextSeed: "a70c36286be722d8111e69e910ce4490005bbf9135b0ce8e7a59f84eee24b88b",
		},
		{
			seed:     "a3b08ccf8cbae4955c02f223be1f97d2bb41d92b7f0c516eb8467a17da1e6057",
			seckey:   "82fba4cc2bc29eef122f116f45d01d82ff488d7ee713f8a95c162a64097239e0",
			pubkey:   "02be1c44351c2e4295e4e9257667b164e2a0e471ecf499084357c13e1b5119b4c2",
			nextSeed: "e9db072fe5817325504174253a056be7b53b512f1e588f576f1f5a82cdcad302",
		},
		{
			seed:     "7048eb8fa93cec992b93dc8e93c5543be34aad05239d4c036cf9e587bbcf7654",
			seckey:   "44c059496aac871ac168bb6889b9dd3decdb9e1fa082442a95fcbca982643425",
			pubkey:   "028868f984547f2febcdd9087a1cc3528929598b1afc9feec0fa62233052bff401",
			nextSeed: "5e9133e83c4add2b0420d485e1dcda5c00e283c6509388ab8ceb583b0485c13b",
		},
		{
			seed:     "6d25375591bbfce7f601fc5eb40e4f3dde2e453dc4bf31595d8ec29e4370cd80",
			seckey:   "d709ceb1a6fb906de506ea091c844ca37c65e52778b8d257d1dd3a942ab367fb",
			pubkey:   "035f0b2cd081f6dd45178d87be62c88b020599292cf77834d8a4dab7a7aad6b1be",
			nextSeed: "8d5579cd702c06c40fb98e1d55121ea0d29f3a6c42f5582b902ac243f29b571a",
		},
		{
			seed:     "7214b4c09f584c5ddff971d469df130b9a3c03e0277e92be159279de39462120",
			seckey:   "5fe4986fa964773041e119d2b6549acb392b2277a72232af75cbfb62c357c1a7",
			pubkey:   "032e039b5885d2d6001b2b5eb4b0af473befa04d2d9fbc4c12ef78f30fe186e487",
			nextSeed: "3a4e8c72921099a0e6a4e7f979df4c8bced63063097835cdfd5ee94548c9c41a",
		},
		{
			seed:     "b13e78392d5446ae304b5fc9d45b85f26996982b2c0c86138afdac8d2ea9016e",
			seckey:   "f784abc2e7f11ee84b4adb72ea4730a6aabe27b09604c8e2b792d8a1a31881ac",
			pubkey:   "02f32b1960c1e61ccc58bb458b8e6fc74a2c37fcb1deb0708251b35e55ba11795e",
			nextSeed: "462efa1bf4f639ffaedb170d6fb8ba363efcb1bdf0c5aef0c75afb59806b8053",
		},
		{
			seed:     "9403bff4240a5999e17e0ab4a645d6942c3a7147c7834e092e461a4580249e6e",
			seckey:   "d495174b8d3f875226b9b939121ec53f9383bd560d34aa5ca3ac6b257512adf4",
			pubkey:   "03f60cefd9bcc6f38487ae082d475c86ee086f0dfed25ff8758c1a9b06862dd0b8",
			nextSeed: "68dd702ea7c7352632876e9dc2333142fce857a542726e402bb480cad364f260",
		},
		{
			seed:     "2665312a3e3628f4df0b9bc6334f530608a9bcdd4d1eef174ecda99f51a6db94",
			seckey:   "1fdc9fbfc6991b9416b3a8385c9942e2db59009aeb2d8de349b73d9f1d389374",
			pubkey:   "029a3b04c75549c8a509fb42a2fa4e8d8361bbe543ee93ccecea90411924f5ab5b",
			nextSeed: "5db72c31d575c332e60f890c7e68d59bd3d0ac53a832e06e821d819476e1f010",
		},
		{
			seed:     "6cb37532c80765b7c07698502a49d69351036f57a45a5143e33c57c236d841ca",
			seckey:   "c87c85a6f482964db7f8c31720981925b1e357a9fdfcc585bc2164fdef1f54d0",
			pubkey:   "02b0f062bdf46066a9a7adb9337a6741ffe95ec26c5652d178dfff88ad302c962d",
			nextSeed: "0deb20ec503b4c678213979fd98018c56f24e9c1ec99af3cd84b43c161a9bb5c",
		},
		{
			seed:     "8654a32fa120bfdb7ca02c487469070eba4b5a81b03763a2185fdf5afd756f3c",
			seckey:   "e2767d788d1c5620f3ef21d57f2d64559ab203c044f0a5f0730b21984e77019c",
			pubkey:   "03251fa5b85a9ada12787234e0ceb3dcc5bd58a49c15ac0749a4238f3bca6d9a1d",
			nextSeed: "36f3ede761aa683813013ffa84e3738b870ce7605e0a958ed4ffb540cd3ea504",
		},
		{
			seed:     "66d1945ceb6ef8014b1b6703cb624f058913e722f15d03225be27cb9d8aabe4a",
			seckey:   "3fcb80eb1d5b91c491408447ac4e221fcb2254c861adbb5a178337c2750b0846",
			pubkey:   "03975043476a76b72b093d684b8a0979d8b246c2f99f16f95760d6d3490c2e37a1",
			nextSeed: "6bcb4819a96508efa7e32ee52b0227ccf5fbe5539687aae931677b24f6d0bbbd",
		},
		{
			seed:     "22c7623bf0e850538329e3e6d9a6f9b1235350824a3feaad2580b7a853550deb",
			seckey:   "5577d4be25f1b44487140a626c8aeca2a77507a1fc4fd466dd3a82234abb6785",
			pubkey:   "0262e2054c89ad173f741e413d12f511a2cf98783c43f18f8dbb6274bdd584a3dc",
			nextSeed: "8bb257a1a17fd2233935b33441d216551d5ff1553d02e4013e03f14962615c16",
		},
		{
			seed:     "a5eebe3469d68c8922a1a8b5a0a2b55293b7ff424240c16feb9f51727f734516",
			seckey:   "c07275582d0681eb07c7b51f0bca0c48c056d571b7b83d84980ab40ac7d7d720",
			pubkey:   "03d80474b8e6002793374a99d884605cf022d216573459b7deb19b6ccb110d286a",
			nextSeed: "d6b780983a63a3e4bcf643ee68b686421079c835a99eeba6962fe41bb355f8da",
		},
		{
			seed:     "479ec3b589b14aa7290b48c2e64072e4e5b15ce395d2072a5a18b0a2cf35f3fd",
			seckey:   "f10e2b7675dfa557d9e3188469f12d3e953c2d46dce006cd177b6ae7f465cfc0",
			pubkey:   "0219d5b487ebdf719a994dcde094072e0227fc23e4cdbc4cce3b9d3e4a4ffe0668",
			nextSeed: "39c5f108e7017e085fe90acfd719420740e57768ac14c94cb020d87e36d06752",
		},
		{
			seed:     "63952334b731ec91d88c54614925576f82e3610d009657368fc866e7b1efbe73",
			seckey:   "0bcbebb39d8fe1cb3eab952c6f701656c234e462b945e2f7d4be2c80b8f2d974",
			pubkey:   "02b58d9eb9595c24438a6ae4a4be4a408c0cd7a3017c3780cba253171cc9e62627",
			nextSeed: "79f654976732106c0e4a97ab3b6d16f343a05ebfcc2e1d679d69d396e6162a77",
		},
		{
			seed:     "256472ee754ef6af096340ab1e161f58e85fb0cc7ae6e6866b9359a1657fa6c1",
			seckey:   "88ba6f6c66fc0ef01c938569c2dd1f05475cb56444f4582d06828e77d54ffbe6",
			pubkey:   "02dac6b246a06ac7c38a63f4f10b1344a8cd6f920a8fd74523bd2f5d4a9a3055b2",
			nextSeed: "387883b86e2acc153aa334518cea48c0c481b573ccaacf17c575623c392f78b2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.seed, func(t *testing.T) {
			seed := MustDecodeHex(tc.seed)
			seckey1 := MustDecodeHex(tc.seckey)
			pubkey1 := MustDecodeHex(tc.pubkey)
			nextSeed1 := MustDecodeHex(tc.nextSeed)
			nextSeed2, pubkey2, seckey2 := DeterministicKeyPairIterator(seed)
			if !bytes.Equal(seckey1, seckey2) {
				t.Error("seckeys don't match")
			}
			if !bytes.Equal(pubkey1, pubkey2) {
				t.Errorf("pubkeys don't match: %s != %s", tc.pubkey, hex.EncodeToString(pubkey2))
			}
			if !bytes.Equal(nextSeed1, nextSeed2) {
				t.Errorf("nextSeeds don't match: %s != %s", tc.nextSeed, hex.EncodeToString(nextSeed2))
			}
		})
	}
}

func TestSecp256k1Hash(t *testing.T) {
	cases := []struct {
		seed string
		hash string
	}{
		{
			seed: "90c56f5b8d78a46fb4cddf6fd9c6d88d6d2d7b0ec35917c7dac12c03b04e444e",
			hash: "a70c36286be722d8111e69e910ce4490005bbf9135b0ce8e7a59f84eee24b88b",
		},
		{
			seed: "a3b08ccf8cbae4955c02f223be1f97d2bb41d92b7f0c516eb8467a17da1e6057",
			hash: "e9db072fe5817325504174253a056be7b53b512f1e588f576f1f5a82cdcad302",
		},
		{
			seed: "7048eb8fa93cec992b93dc8e93c5543be34aad05239d4c036cf9e587bbcf7654",
			hash: "5e9133e83c4add2b0420d485e1dcda5c00e283c6509388ab8ceb583b0485c13b",
		},
		{
			seed: "6d25375591bbfce7f601fc5eb40e4f3dde2e453dc4bf31595d8ec29e4370cd80",
			hash: "8d5579cd702c06c40fb98e1d55121ea0d29f3a6c42f5582b902ac243f29b571a",
		},
		{
			seed: "7214b4c09f584c5ddff971d469df130b9a3c03e0277e92be159279de39462120",
			hash: "3a4e8c72921099a0e6a4e7f979df4c8bced63063097835cdfd5ee94548c9c41a",
		},
		{
			seed: "b13e78392d5446ae304b5fc9d45b85f26996982b2c0c86138afdac8d2ea9016e",
			hash: "462efa1bf4f639ffaedb170d6fb8ba363efcb1bdf0c5aef0c75afb59806b8053",
		},
		{
			seed: "9403bff4240a5999e17e0ab4a645d6942c3a7147c7834e092e461a4580249e6e",
			hash: "68dd702ea7c7352632876e9dc2333142fce857a542726e402bb480cad364f260",
		},
		{
			seed: "2665312a3e3628f4df0b9bc6334f530608a9bcdd4d1eef174ecda99f51a6db94",
			hash: "5db72c31d575c332e60f890c7e68d59bd3d0ac53a832e06e821d819476e1f010",
		},
		{
			seed: "6cb37532c80765b7c07698502a49d69351036f57a45a5143e33c57c236d841ca",
			hash: "0deb20ec503b4c678213979fd98018c56f24e9c1ec99af3cd84b43c161a9bb5c",
		},
		{
			seed: "8654a32fa120bfdb7ca02c487469070eba4b5a81b03763a2185fdf5afd756f3c",
			hash: "36f3ede761aa683813013ffa84e3738b870ce7605e0a958ed4ffb540cd3ea504",
		},
		{
			seed: "66d1945ceb6ef8014b1b6703cb624f058913e722f15d03225be27cb9d8aabe4a",
			hash: "6bcb4819a96508efa7e32ee52b0227ccf5fbe5539687aae931677b24f6d0bbbd",
		},
		{
			seed: "22c7623bf0e850538329e3e6d9a6f9b1235350824a3feaad2580b7a853550deb",
			hash: "8bb257a1a17fd2233935b33441d216551d5ff1553d02e4013e03f14962615c16",
		},
		{
			seed: "a5eebe3469d68c8922a1a8b5a0a2b55293b7ff424240c16feb9f51727f734516",
			hash: "d6b780983a63a3e4bcf643ee68b686421079c835a99eeba6962fe41bb355f8da",
		},
		{
			seed: "479ec3b589b14aa7290b48c2e64072e4e5b15ce395d2072a5a18b0a2cf35f3fd",
			hash: "39c5f108e7017e085fe90acfd719420740e57768ac14c94cb020d87e36d06752",
		},
		{
			seed: "63952334b731ec91d88c54614925576f82e3610d009657368fc866e7b1efbe73",
			hash: "79f654976732106c0e4a97ab3b6d16f343a05ebfcc2e1d679d69d396e6162a77",
		},
		{
			seed: "256472ee754ef6af096340ab1e161f58e85fb0cc7ae6e6866b9359a1657fa6c1",
			hash: "387883b86e2acc153aa334518cea48c0c481b573ccaacf17c575623c392f78b2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.seed, func(t *testing.T) {
			seed := MustDecodeHex(tc.seed)
			if len(seed) != 32 {
				t.Fatal("expected seed to be 32 bytes")
			}
			hash := Secp256k1Hash(seed)
			if len(hash) != 32 {
				t.Fatal("expected hash to be 32 bytes")
			}
			if !bytes.Equal(MustDecodeHex(tc.hash), hash) {
				t.Fatal("secp256k1Hash does not match")
			}
		})
	}
}

func TestSecp256k1Equal(t *testing.T) {
	for i := 0; i < 64; i++ {
		seed := RandByte(128)

		hash1 := Secp256k1Hash(seed)
		hash2, _, _ := DeterministicKeyPairIterator(seed)

		if !bytes.Equal(hash1, hash2) {
			t.Fail()
		}
	}
}

func TestECDH(t *testing.T) {
	pubkey1, seckey1 := GenerateKeyPair()
	pubkey2, seckey2 := GenerateKeyPair()

	puba := ECDH(pubkey1, seckey2)
	pubb := ECDH(pubkey2, seckey1)

	if puba == nil {
		t.Fail()
	}

	if pubb == nil {
		t.Fail()
	}

	if !bytes.Equal(puba, pubb) {
		t.Fail()
	}
}

func TestECDH2(t *testing.T) {
	for i := 0; i < 16*1024; i++ {

		pubkey1, seckey1 := GenerateKeyPair()
		pubkey2, seckey2 := GenerateKeyPair()

		puba := ECDH(pubkey1, seckey2)
		pubb := ECDH(pubkey2, seckey1)

		if puba == nil {
			t.Fail()
		}

		if pubb == nil {
			t.Fail()
		}

		if !bytes.Equal(puba, pubb) {
			t.Fail()
		}
	}
}

func TestDeterministicKeyPairIteratorStepRandom(t *testing.T) {
	n := 32 * 1024
	for i := 0; i < n; i++ {
		seed := RandByte(32)

		pubkey1, seckey1 := deterministicKeyPairIteratorStep(seed)

		if seckey1 == nil {
			t.Fail()
		}

		if pubkey1 == nil {
			t.Fail()
		}

		if VerifyPubkey(pubkey1) != 1 {
			seedHex := hex.EncodeToString(seed)
			seckeyHex := hex.EncodeToString(seckey1)
			t.Logf("seed=%s", seedHex)
			t.Logf("seckey=%s", seckeyHex)
			t.Errorf("GenerateKeyPair, generates key that fails validation, run=%d", i)
		}
	}
}

// problematic seckeys
// See: https://github.com/piotrnar/gocoin/issues/15
// For additional information on the origin of these test vectors
var abnormalSecKeys = []struct {
	sec  string
	pub  string
	ecdh []string
}{
	{
		sec: "08efb79385c9a8b0d1c6f5f6511be0c6f6c2902963d874a3a4bacc18802528d3",
		pub: "03c74332d6094b1f603d4902fc6b1aa09fb3ef81f3015a4000cc0077ff70543c16",
		ecdh: []string{
			"",
			"02e72655a3adf8308a078ee6fe948cf6baf95ef626b1e1fe6e434c737c7c2fef4e",
			"03222fe59be5a69c38364dd313bd077b8b1c2216804a4a727e0078b3c77778bc45",
			"021096aa98231eaa949542be029a1f3a93815e05e243c69e73d7449d719ff5d76d",
		},
	},
	{
		sec: "78298d9ecdc0640c9ae6883201a53f4518055442642024d23c45858f45d0c3e6",
		pub: "02fa3e6e0b1fb76e26dffe7b1e01fd02677fedfed23a59000092c706b04214bee3",
		ecdh: []string{
			"02e72655a3adf8308a078ee6fe948cf6baf95ef626b1e1fe6e434c737c7c2fef4e",
			"",
			"025617125b44ded369deed72f833535d56a3ed035afc44ff64fb7c65986f6ea2a5",
			"03849b3f906180cf27c161045e9da551a44476b0d4f7f29d668ba17569953d0a11",
		},
	},
	{
		sec: "04e04fe65bfa6ded50a12769a3bd83d7351b2dbff08c9bac14662b23a3294b9e",
		pub: "034f25c9400dd0f87a9c420b35b5a157d21caa086ef8fa00015bc3c8ab73a1cc4c",
		ecdh: []string{
			"03222fe59be5a69c38364dd313bd077b8b1c2216804a4a727e0078b3c77778bc45",
			"025617125b44ded369deed72f833535d56a3ed035afc44ff64fb7c65986f6ea2a5",
			"",
			"03fd41f8d279e2df640f17aef31c258a0a9aa6ddcaf4c4bc80f71dccff576b630c",
		},
	},
	{
		sec: "2f5141f1b75747996c5de77c911dae062d16ae48799052c04ead20ccd5afa113",
		pub: "03fe58baefc491a9dcf0939ab6252f81f6d9515105bd89c000bb7f2a694e8a8b72",
		ecdh: []string{
			"021096aa98231eaa949542be029a1f3a93815e05e243c69e73d7449d719ff5d76d",
			"03849b3f906180cf27c161045e9da551a44476b0d4f7f29d668ba17569953d0a11",
			"03fd41f8d279e2df640f17aef31c258a0a9aa6ddcaf4c4bc80f71dccff576b630c",
			"",
		},
	},
}

func TestAbnormalKeys2(t *testing.T) {
	for _, tc := range abnormalSecKeys {
		seckey1, err := hex.DecodeString(tc.sec)
		if err != nil {
			t.Error(err)
		}

		pubkey1 := PubkeyFromSeckey(seckey1)
		if pubkey1 == nil {
			t.Fail()
		}

		if seckey1 == nil {
			t.Fail()
		}

		if pubkey1 == nil {
			t.Fail()
		}

		if VerifyPubkey(pubkey1) != 1 {
			t.Error("generates key that fails validation")
		}

		hpubkey1 := hex.EncodeToString(pubkey1)
		if hpubkey1 != tc.pub {
			t.Errorf("pubkey does not match %s != %s", hpubkey1, tc.pub)
		}
	}
}

func TestAbnormalKeys3(t *testing.T) {
	// ECDH test
	for i, tc := range abnormalSecKeys {
		seckey1, err := hex.DecodeString(tc.sec)
		if err != nil {
			t.Error(err)
		}

		pubkey1 := PubkeyFromSeckey(seckey1)

		if pubkey1 == nil {
			t.Error("pubkey1 nil")
		}

		if hex.EncodeToString(pubkey1) != tc.pub {
			t.Error("pubkey1 does not match")
		}

		for j, tc2 := range abnormalSecKeys {
			if i == j {
				continue
			}

			seckey2, err := hex.DecodeString(tc2.sec)
			if err != nil {
				t.Error(err)
			}
			pubkey2 := PubkeyFromSeckey(seckey2)

			if pubkey2 == nil {
				t.Error("pubkey2 nil")
			}

			if hex.EncodeToString(pubkey2) != tc2.pub {
				t.Error("pubkey2 does not match")
			}

			puba := ECDH(pubkey1, seckey2)
			pubb := ECDH(pubkey2, seckey1)

			if puba == nil {
				t.Fail()
			}

			if pubb == nil {
				t.Fail()
			}

			if !bytes.Equal(puba, pubb) {
				t.Error("recovered ecdh keys do not match")
			}

			hpuba := hex.EncodeToString(puba)
			if hpuba != tc.ecdh[j] {
				t.Errorf("ecdh does not match %d,%d %s != %s", i, j, hpuba, tc.ecdh[j])
			}
		}

	}
}

func TestDeterministicKeyPairIterator2(t *testing.T) {
	cases := []struct {
		seed string
		sec  string
		pub  string
		n    int
	}{
		{
			seed: "67c53b28b8c7b06be53b490c28c0a3b77724b5c31c4bf12b71cd44c6bb4586f3",
			sec:  "68c751a58f48d656e4d3ec31f6c1016e6e36583ac2f63129f576b29e764469b5",
			pub:  "02c32556c48bfe944e4b8f6ecb6c884112c71a468247d338cbbdc9c561ab7c6d3d",
			n:    1,
		},
		{
			seed: "38363534613332666131323062666462376361303263343837343639303730656261346235613831623033373633613231383566646635616664373536663363",
			sec:  "10ba0325f1b8633ca463542950b5cd5f97753a9829ba23477c584e7aee9cfbd5",
			pub:  "0249964ac7e3fe1b2c182a2f10abe031784e374cc0c665a63bc76cc009a05bc7c6",
			n:    1024,
		},
	}

	for _, tc := range cases {
		t.Run(tc.seed, func(t *testing.T) {
			seed, err := hex.DecodeString(tc.seed)
			if err != nil {
				t.Fatalf("hex.DecodeString(%s) failed: %v", tc.seed, err)
				return
			}

			var pub []byte
			var sec []byte
			for i := 0; i < tc.n; i++ {
				seed, pub, sec = DeterministicKeyPairIterator(seed)
			}

			hsec := hex.EncodeToString(sec)
			if hsec != tc.sec {
				t.Fatalf("secret key does not match %s != %s", hsec, tc.sec)
			}

			hpub := hex.EncodeToString(pub)
			if hpub != tc.pub {
				t.Fatalf("public key does not match %s != %s", hpub, tc.pub)
			}
		})
	}
}
