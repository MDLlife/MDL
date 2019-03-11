package cipher

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MDLlife/MDL/src/cipher/ripemd160"
)

func freshSumRipemd160(t *testing.T, b []byte) Ripemd160 {
	sh := ripemd160.New()
	_, err := sh.Write(b)
	require.NoError(t, err)
	h := Ripemd160{}
	h.MustSet(sh.Sum(nil))
	return h
}

func freshSumSHA256(t *testing.T, b []byte) SHA256 {
	sh := sha256.New()
	_, err := sh.Write(b)
	require.NoError(t, err)
	h := SHA256{}
	h.MustSet(sh.Sum(nil))
	return h
}

func randBytes(t *testing.T, n int) []byte {
	b := make([]byte, n)
	x, err := rand.Read(b)
	require.Equal(t, n, x)
	require.Nil(t, err)
	return b
}

func TestHashRipemd160(t *testing.T) {
	require.NotPanics(t, func() { HashRipemd160(randBytes(t, 128)) })
	r := HashRipemd160(randBytes(t, 160))
	require.NotEqual(t, r, Ripemd160{})
	// 2nd hash should not be affected by previous
	b := randBytes(t, 256)
	r2 := HashRipemd160(b)
	require.NotEqual(t, r2, Ripemd160{})
	require.Equal(t, r2, freshSumRipemd160(t, b))
}

func TestRipemd160MustSet(t *testing.T) {
	h := Ripemd160{}
	require.Panics(t, func() {
		h.MustSet(randBytes(t, 21))
	})
	require.Panics(t, func() {
		h.MustSet(randBytes(t, 100))
	})
	require.Panics(t, func() {
		h.MustSet(randBytes(t, 19))
	})
	require.Panics(t, func() {
		h.MustSet(randBytes(t, 0))
	})
	require.NotPanics(t, func() {
		h.MustSet(randBytes(t, 20))
	})
	b := randBytes(t, 20)
	h.MustSet(b)
	require.True(t, bytes.Equal(h[:], b))
}

func TestRipemd160Set(t *testing.T) {
	h := Ripemd160{}
	err := h.Set(randBytes(t, 21))
	require.Equal(t, errors.New("Invalid ripemd160 length"), err)
	err = h.Set(randBytes(t, 100))
	require.Equal(t, errors.New("Invalid ripemd160 length"), err)
	err = h.Set(randBytes(t, 19))
	require.Equal(t, errors.New("Invalid ripemd160 length"), err)
	err = h.Set(randBytes(t, 0))
	require.Equal(t, errors.New("Invalid ripemd160 length"), err)

	b := randBytes(t, 20)
	err = h.Set(b)
	require.NoError(t, err)
	require.True(t, bytes.Equal(h[:], b))
}

func TestSHA256MustSet(t *testing.T) {
	h := SHA256{}
	require.Panics(t, func() {
		h.MustSet(randBytes(t, 33))
	})
	require.Panics(t, func() {
		h.MustSet(randBytes(t, 100))
	})
	require.Panics(t, func() {
		h.MustSet(randBytes(t, 31))
	})
	require.Panics(t, func() {
		h.MustSet(randBytes(t, 0))
	})
	require.NotPanics(t, func() {
		h.MustSet(randBytes(t, 32))
	})
	b := randBytes(t, 32)
	h.MustSet(b)
	require.True(t, bytes.Equal(h[:], b))
}

func TestRipemd160FromBytes(t *testing.T) {
	b := randBytes(t, 20)
	h, err := Ripemd160FromBytes(b)
	require.NoError(t, err)
	require.True(t, bytes.Equal(b[:], h[:]))

	b = randBytes(t, 19)
	_, err = Ripemd160FromBytes(b)
	require.Equal(t, errors.New("Invalid ripemd160 length"), err)

	b = randBytes(t, 21)
	_, err = Ripemd160FromBytes(b)
	require.Equal(t, errors.New("Invalid ripemd160 length"), err)

	_, err = Ripemd160FromBytes(nil)
	require.Equal(t, errors.New("Invalid ripemd160 length"), err)
}

func TestMustRipemd160FromBytes(t *testing.T) {
	b := randBytes(t, 20)
	h := MustRipemd160FromBytes(b)
	require.True(t, bytes.Equal(b[:], h[:]))

	b = randBytes(t, 19)
	require.Panics(t, func() {
		MustRipemd160FromBytes(b)
	})

	b = randBytes(t, 21)
	require.Panics(t, func() {
		MustRipemd160FromBytes(b)
	})

	require.Panics(t, func() {
		MustRipemd160FromBytes(nil)
	})
}

func TestSHA256Set(t *testing.T) {
	h := SHA256{}
	err := h.Set(randBytes(t, 33))
	require.Equal(t, errors.New("Invalid sha256 length"), err)
	err = h.Set(randBytes(t, 100))
	require.Equal(t, errors.New("Invalid sha256 length"), err)
	err = h.Set(randBytes(t, 31))
	require.Equal(t, errors.New("Invalid sha256 length"), err)
	err = h.Set(randBytes(t, 0))
	require.Equal(t, errors.New("Invalid sha256 length"), err)

	b := randBytes(t, 32)
	err = h.Set(b)
	require.NoError(t, err)
	require.True(t, bytes.Equal(h[:], b))
}

func TestSHA256Hex(t *testing.T) {
	h := SHA256{}
	h.MustSet(randBytes(t, 32))
	s := h.Hex()
	h2, err := SHA256FromHex(s)
	require.Nil(t, err)
	require.Equal(t, h, h2)
	require.Equal(t, h2.Hex(), s)
}

func TestSHA256KnownValue(t *testing.T) {
	vals := []struct {
		input  string
		output string
	}{
		// These values are generated by
		// echo -n input | sha256sum
		{
			"mdl",
			"d3c3c54797643905c5cc97f7da4717058dbe6ad183ef1586104cadd197ca47c6",
		},
		{
			"hello world",
			"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			"hello world asd awd awd awdapodawpokawpod ",
			"99d71f95cafe05ea2dddebc35b6083bd5af0e44850c9dc5139b4476c99950be4",
		},
	}
	for _, io := range vals {
		require.Equal(t, io.output, SumSHA256([]byte(io.input)).Hex())
	}
}

func TestSumSHA256(t *testing.T) {
	b := randBytes(t, 256)
	h1 := SumSHA256(b)
	require.NotEqual(t, h1, SHA256{})
	// A second call to Sum should not be influenced by the original
	c := randBytes(t, 256)
	h2 := SumSHA256(c)
	require.NotEqual(t, h2, SHA256{})
	require.Equal(t, h2, freshSumSHA256(t, c))
}

func TestSHA256FromHex(t *testing.T) {
	// Invalid hex hash
	_, err := SHA256FromHex("cawcd")
	require.NotNil(t, err)

	// Truncated hex hash
	h := SumSHA256(randBytes(t, 128))
	_, err = SHA256FromHex(hex.EncodeToString(h[:len(h)/2]))
	require.NotNil(t, err)

	// Valid hex hash
	h2, err := SHA256FromHex(hex.EncodeToString(h[:]))
	require.Equal(t, h, h2)
	require.Nil(t, err)
}

func TestMustSHA256FromHex(t *testing.T) {
	// Invalid hex hash
	require.Panics(t, func() { MustSHA256FromHex("cawcd") })

	// Truncated hex hash
	h := SumSHA256(randBytes(t, 128))
	require.Panics(t, func() {
		MustSHA256FromHex(hex.EncodeToString(h[:len(h)/2]))
	})

	// Valid hex hash
	h2 := MustSHA256FromHex(hex.EncodeToString(h[:]))
	require.Equal(t, h, h2)
}

func TestSHA256FromBytes(t *testing.T) {
	b := randBytes(t, 32)
	h, err := SHA256FromBytes(b)
	require.NoError(t, err)
	require.True(t, bytes.Equal(b[:], h[:]))

	b = randBytes(t, 31)
	_, err = SHA256FromBytes(b)
	require.Equal(t, errors.New("Invalid sha256 length"), err)

	b = randBytes(t, 33)
	_, err = SHA256FromBytes(b)
	require.Equal(t, errors.New("Invalid sha256 length"), err)

	_, err = SHA256FromBytes(nil)
	require.Equal(t, errors.New("Invalid sha256 length"), err)
}

func TestMustSHA256FromBytes(t *testing.T) {
	b := randBytes(t, 32)
	h := MustSHA256FromBytes(b)
	require.True(t, bytes.Equal(b[:], h[:]))

	b = randBytes(t, 31)
	require.Panics(t, func() {
		MustSHA256FromBytes(b)
	})

	b = randBytes(t, 33)
	require.Panics(t, func() {
		MustSHA256FromBytes(b)
	})

	require.Panics(t, func() {
		MustSHA256FromBytes(nil)
	})
}

func TestDoubleSHA256(t *testing.T) {
	b := randBytes(t, 128)
	h := DoubleSHA256(b)
	require.NotEqual(t, h, SHA256{})
	require.NotEqual(t, h, freshSumSHA256(t, b))
}

func TestAddSHA256(t *testing.T) {
	b := randBytes(t, 128)
	h := SumSHA256(b)
	c := randBytes(t, 64)
	i := SumSHA256(c)
	add := AddSHA256(h, i)
	require.NotEqual(t, add, SHA256{})
	require.NotEqual(t, add, h)
	require.NotEqual(t, add, i)
	require.Equal(t, add, SumSHA256(append(h[:], i[:]...)))
}

func TestXorSHA256(t *testing.T) {
	b := randBytes(t, 128)
	c := randBytes(t, 128)
	h := SumSHA256(b)
	i := SumSHA256(c)
	require.NotEqual(t, h.Xor(i), h)
	require.NotEqual(t, h.Xor(i), i)
	require.NotEqual(t, h.Xor(i), SHA256{})
	require.Equal(t, h.Xor(i), i.Xor(h))
}

func TestSHA256Null(t *testing.T) {
	var x SHA256
	require.True(t, x.Null())

	b := randBytes(t, 128)
	x = SumSHA256(b)

	require.False(t, x.Null())
}

func TestNextPowerOfTwo(t *testing.T) {
	inputs := [][]uint64{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{8, 8},
		{14, 16},
		{16, 16},
		{17, 32},
		{43345, 65536},
		{65535, 65536},
		{35657, 65536},
		{65536, 65536},
		{65537, 131072},
	}
	for _, i := range inputs {
		require.Equal(t, nextPowerOfTwo(i[0]), i[1])
	}
	for i := uint64(2); i < 10000; i++ {
		p := nextPowerOfTwo(i)
		require.Equal(t, p%2, uint64(0))
		require.True(t, p >= i)
	}
}

func TestMerkle(t *testing.T) {
	h := SumSHA256(randBytes(t, 128))
	// Single hash input returns hash
	require.Equal(t, Merkle([]SHA256{h}), h)
	h2 := SumSHA256(randBytes(t, 128))
	// 2 hashes should be AddSHA256 of them
	require.Equal(t, Merkle([]SHA256{h, h2}), AddSHA256(h, h2))
	// 3 hashes should be Add(Add())
	h3 := SumSHA256(randBytes(t, 128))
	out := AddSHA256(AddSHA256(h, h2), AddSHA256(h3, SHA256{}))
	require.Equal(t, Merkle([]SHA256{h, h2, h3}), out)
	// 4 hashes should be Add(Add())
	h4 := SumSHA256(randBytes(t, 128))
	out = AddSHA256(AddSHA256(h, h2), AddSHA256(h3, h4))
	require.Equal(t, Merkle([]SHA256{h, h2, h3, h4}), out)
	// 5 hashes
	h5 := SumSHA256(randBytes(t, 128))
	out = AddSHA256(AddSHA256(h, h2), AddSHA256(h3, h4))
	out = AddSHA256(out, AddSHA256(AddSHA256(h5, SHA256{}),
		AddSHA256(SHA256{}, SHA256{})))
	require.Equal(t, Merkle([]SHA256{h, h2, h3, h4, h5}), out)
}
