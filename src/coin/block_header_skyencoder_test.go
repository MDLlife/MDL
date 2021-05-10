// Code generated by github.com/MDLlife/skyencoder. DO NOT EDIT.
package coin

import (
	"bytes"
	"fmt"
	mathrand "math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

    "github.com/skycoin/encodertest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/MDLlife/MDL/src/cipher/encoder"
)

func newEmptyBlockHeaderForEncodeTest() *BlockHeader {
	var obj BlockHeader
	return &obj
}

func newRandomBlockHeaderForEncodeTest(t *testing.T, rand *mathrand.Rand) *BlockHeader {
	var obj BlockHeader
	err := encodertest.PopulateRandom(&obj, rand, encodertest.PopulateRandomOptions{
		MaxRandLen: 4,
		MinRandLen: 1,
	})
	if err != nil {
		t.Fatalf("encodertest.PopulateRandom failed: %v", err)
	}
	return &obj
}

func newRandomZeroLenBlockHeaderForEncodeTest(t *testing.T, rand *mathrand.Rand) *BlockHeader {
	var obj BlockHeader
	err := encodertest.PopulateRandom(&obj, rand, encodertest.PopulateRandomOptions{
		MaxRandLen:    0,
		MinRandLen:    0,
		EmptySliceNil: false,
		EmptyMapNil:   false,
	})
	if err != nil {
		t.Fatalf("encodertest.PopulateRandom failed: %v", err)
	}
	return &obj
}

func newRandomZeroLenNilBlockHeaderForEncodeTest(t *testing.T, rand *mathrand.Rand) *BlockHeader {
	var obj BlockHeader
	err := encodertest.PopulateRandom(&obj, rand, encodertest.PopulateRandomOptions{
		MaxRandLen:    0,
		MinRandLen:    0,
		EmptySliceNil: true,
		EmptyMapNil:   true,
	})
	if err != nil {
		t.Fatalf("encodertest.PopulateRandom failed: %v", err)
	}
	return &obj
}

func testSkyencoderBlockHeader(t *testing.T, obj *BlockHeader) {
	isEncodableField := func(f reflect.StructField) bool {
		// Skip unexported fields
		if f.PkgPath != "" {
			return false
		}

		// Skip fields disabled with and enc:"- struct tag
		tag := f.Tag.Get("enc")
		return !strings.HasPrefix(tag, "-,") && tag != "-"
	}

	hasOmitEmptyField := func(obj interface{}) bool {
		v := reflect.ValueOf(obj)
		switch v.Kind() {
		case reflect.Ptr:
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			t := v.Type()
			n := v.NumField()
			f := t.Field(n - 1)
			tag := f.Tag.Get("enc")
			return isEncodableField(f) && strings.Contains(tag, ",omitempty")
		default:
			return false
		}
	}

	// returns the number of bytes encoded by an omitempty field on a given object
	omitEmptyLen := func(obj interface{}) uint64 {
		if !hasOmitEmptyField(obj) {
			return 0
		}

		v := reflect.ValueOf(obj)
		switch v.Kind() {
		case reflect.Ptr:
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			n := v.NumField()
			f := v.Field(n - 1)
			if f.Len() == 0 {
				return 0
			}
			return uint64(4 + f.Len())

		default:
			return 0
		}
	}

	// encodeSize

	n1 := encoder.Size(obj)
	n2 := encodeSizeBlockHeader(obj)

	if uint64(n1) != n2 {
		t.Fatalf("encoder.Size() != encodeSizeBlockHeader() (%d != %d)", n1, n2)
	}

	// Encode

	// encoder.Serialize
	data1 := encoder.Serialize(obj)

	// Encode
	data2, err := encodeBlockHeader(obj)
	if err != nil {
		t.Fatalf("encodeBlockHeader failed: %v", err)
	}
	if uint64(len(data2)) != n2 {
		t.Fatal("encodeBlockHeader produced bytes of unexpected length")
	}
	if len(data1) != len(data2) {
		t.Fatalf("len(encoder.Serialize()) != len(encodeBlockHeader()) (%d != %d)", len(data1), len(data2))
	}

	// EncodeToBuffer
	data3 := make([]byte, n2+5)
	if err := encodeBlockHeaderToBuffer(data3, obj); err != nil {
		t.Fatalf("encodeBlockHeaderToBuffer failed: %v", err)
	}

	if !bytes.Equal(data1, data2) {
		t.Fatal("encoder.Serialize() != encode[1]s()")
	}

	// Decode

	// encoder.DeserializeRaw
	var obj2 BlockHeader
	if n, err := encoder.DeserializeRaw(data1, &obj2); err != nil {
		t.Fatalf("encoder.DeserializeRaw failed: %v", err)
	} else if n != uint64(len(data1)) {
		t.Fatalf("encoder.DeserializeRaw failed: %v", encoder.ErrRemainingBytes)
	}
	if !cmp.Equal(*obj, obj2, cmpopts.EquateEmpty(), encodertest.IgnoreAllUnexported()) {
		t.Fatal("encoder.DeserializeRaw result wrong")
	}

	// Decode
	var obj3 BlockHeader
	if n, err := decodeBlockHeader(data2, &obj3); err != nil {
		t.Fatalf("decodeBlockHeader failed: %v", err)
	} else if n != uint64(len(data2)) {
		t.Fatalf("decodeBlockHeader bytes read length should be %d, is %d", len(data2), n)
	}
	if !cmp.Equal(obj2, obj3, cmpopts.EquateEmpty(), encodertest.IgnoreAllUnexported()) {
		t.Fatal("encoder.DeserializeRaw() != decodeBlockHeader()")
	}

	// Decode, excess buffer
	var obj4 BlockHeader
	n, err := decodeBlockHeader(data3, &obj4)
	if err != nil {
		t.Fatalf("decodeBlockHeader failed: %v", err)
	}

	if hasOmitEmptyField(&obj4) && omitEmptyLen(&obj4) == 0 {
		// 4 bytes read for the omitEmpty length, which should be zero (see the 5 bytes added above)
		if n != n2+4 {
			t.Fatalf("decodeBlockHeader bytes read length should be %d, is %d", n2+4, n)
		}
	} else {
		if n != n2 {
			t.Fatalf("decodeBlockHeader bytes read length should be %d, is %d", n2, n)
		}
	}
	if !cmp.Equal(obj2, obj4, cmpopts.EquateEmpty(), encodertest.IgnoreAllUnexported()) {
		t.Fatal("encoder.DeserializeRaw() != decodeBlockHeader()")
	}

	// DecodeExact
	var obj5 BlockHeader
	if err := decodeBlockHeaderExact(data2, &obj5); err != nil {
		t.Fatalf("decodeBlockHeader failed: %v", err)
	}
	if !cmp.Equal(obj2, obj5, cmpopts.EquateEmpty(), encodertest.IgnoreAllUnexported()) {
		t.Fatal("encoder.DeserializeRaw() != decodeBlockHeader()")
	}

	// Check that the bytes read value is correct when providing an extended buffer
	if !hasOmitEmptyField(&obj3) || omitEmptyLen(&obj3) > 0 {
		padding := []byte{0xFF, 0xFE, 0xFD, 0xFC}
		data4 := append(data2[:], padding...)
		if n, err := decodeBlockHeader(data4, &obj3); err != nil {
			t.Fatalf("decodeBlockHeader failed: %v", err)
		} else if n != uint64(len(data2)) {
			t.Fatalf("decodeBlockHeader bytes read length should be %d, is %d", len(data2), n)
		}
	}
}

func TestSkyencoderBlockHeader(t *testing.T) {
	rand := mathrand.New(mathrand.NewSource(time.Now().Unix()))

	type testCase struct {
		name string
		obj  *BlockHeader
	}

	cases := []testCase{
		{
			name: "empty object",
			obj:  newEmptyBlockHeaderForEncodeTest(),
		},
	}

	nRandom := 10

	for i := 0; i < nRandom; i++ {
		cases = append(cases, testCase{
			name: fmt.Sprintf("randomly populated object %d", i),
			obj:  newRandomBlockHeaderForEncodeTest(t, rand),
		})
		cases = append(cases, testCase{
			name: fmt.Sprintf("randomly populated object %d with zero length variable length contents", i),
			obj:  newRandomZeroLenBlockHeaderForEncodeTest(t, rand),
		})
		cases = append(cases, testCase{
			name: fmt.Sprintf("randomly populated object %d with zero length variable length contents set to nil", i),
			obj:  newRandomZeroLenNilBlockHeaderForEncodeTest(t, rand),
		})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testSkyencoderBlockHeader(t, tc.obj)
		})
	}
}

func decodeBlockHeaderExpectError(t *testing.T, buf []byte, expectedErr error) {
	var obj BlockHeader
	if _, err := decodeBlockHeader(buf, &obj); err == nil {
		t.Fatal("decodeBlockHeader: expected error, got nil")
	} else if err != expectedErr {
		t.Fatalf("decodeBlockHeader: expected error %q, got %q", expectedErr, err)
	}
}

func decodeBlockHeaderExactExpectError(t *testing.T, buf []byte, expectedErr error) {
	var obj BlockHeader
	if err := decodeBlockHeaderExact(buf, &obj); err == nil {
		t.Fatal("decodeBlockHeaderExact: expected error, got nil")
	} else if err != expectedErr {
		t.Fatalf("decodeBlockHeaderExact: expected error %q, got %q", expectedErr, err)
	}
}

func testSkyencoderBlockHeaderDecodeErrors(t *testing.T, k int, tag string, obj *BlockHeader) {
	isEncodableField := func(f reflect.StructField) bool {
		// Skip unexported fields
		if f.PkgPath != "" {
			return false
		}

		// Skip fields disabled with and enc:"- struct tag
		tag := f.Tag.Get("enc")
		return !strings.HasPrefix(tag, "-,") && tag != "-"
	}

	numEncodableFields := func(obj interface{}) int {
		v := reflect.ValueOf(obj)
		switch v.Kind() {
		case reflect.Ptr:
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			t := v.Type()

			n := 0
			for i := 0; i < v.NumField(); i++ {
				f := t.Field(i)
				if !isEncodableField(f) {
					continue
				}
				n++
			}
			return n
		default:
			return 0
		}
	}

	hasOmitEmptyField := func(obj interface{}) bool {
		v := reflect.ValueOf(obj)
		switch v.Kind() {
		case reflect.Ptr:
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			t := v.Type()
			n := v.NumField()
			f := t.Field(n - 1)
			tag := f.Tag.Get("enc")
			return isEncodableField(f) && strings.Contains(tag, ",omitempty")
		default:
			return false
		}
	}

	// returns the number of bytes encoded by an omitempty field on a given object
	omitEmptyLen := func(obj interface{}) uint64 {
		if !hasOmitEmptyField(obj) {
			return 0
		}

		v := reflect.ValueOf(obj)
		switch v.Kind() {
		case reflect.Ptr:
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			n := v.NumField()
			f := v.Field(n - 1)
			if f.Len() == 0 {
				return 0
			}
			return uint64(4 + f.Len())

		default:
			return 0
		}
	}

	n := encodeSizeBlockHeader(obj)
	buf, err := encodeBlockHeader(obj)
	if err != nil {
		t.Fatalf("encodeBlockHeader failed: %v", err)
	}

	// A nil buffer cannot decode, unless the object is a struct with a single omitempty field
	if hasOmitEmptyField(obj) && numEncodableFields(obj) > 1 {
		t.Run(fmt.Sprintf("%d %s buffer underflow nil", k, tag), func(t *testing.T) {
			decodeBlockHeaderExpectError(t, nil, encoder.ErrBufferUnderflow)
		})

		t.Run(fmt.Sprintf("%d %s exact buffer underflow nil", k, tag), func(t *testing.T) {
			decodeBlockHeaderExactExpectError(t, nil, encoder.ErrBufferUnderflow)
		})
	}

	// Test all possible truncations of the encoded byte array, but skip
	// a truncation that would be valid where omitempty is removed
	skipN := n - omitEmptyLen(obj)
	for i := uint64(0); i < n; i++ {
		if i == skipN {
			continue
		}

		t.Run(fmt.Sprintf("%d %s buffer underflow bytes=%d", k, tag, i), func(t *testing.T) {
			decodeBlockHeaderExpectError(t, buf[:i], encoder.ErrBufferUnderflow)
		})

		t.Run(fmt.Sprintf("%d %s exact buffer underflow bytes=%d", k, tag, i), func(t *testing.T) {
			decodeBlockHeaderExactExpectError(t, buf[:i], encoder.ErrBufferUnderflow)
		})
	}

	// Append 5 bytes for omit empty with a 0 length prefix, to cause an ErrRemainingBytes.
	// If only 1 byte is appended, the decoder will try to read the 4-byte length prefix,
	// and return an ErrBufferUnderflow instead
	if hasOmitEmptyField(obj) {
		buf = append(buf, []byte{0, 0, 0, 0, 0}...)
	} else {
		buf = append(buf, 0)
	}

	t.Run(fmt.Sprintf("%d %s exact buffer remaining bytes", k, tag), func(t *testing.T) {
		decodeBlockHeaderExactExpectError(t, buf, encoder.ErrRemainingBytes)
	})
}

func TestSkyencoderBlockHeaderDecodeErrors(t *testing.T) {
	rand := mathrand.New(mathrand.NewSource(time.Now().Unix()))
	n := 10

	for i := 0; i < n; i++ {
		emptyObj := newEmptyBlockHeaderForEncodeTest()
		fullObj := newRandomBlockHeaderForEncodeTest(t, rand)
		testSkyencoderBlockHeaderDecodeErrors(t, i, "empty", emptyObj)
		testSkyencoderBlockHeaderDecodeErrors(t, i, "full", fullObj)
	}
}
