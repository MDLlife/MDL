// build ignore

package coin

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/MDLlife/MDL/src/cipher"
	"github.com/MDLlife/MDL/src/testutil"
)

var (
	genPublic, genSecret        = cipher.GenerateKeyPair()
	genAddress                  = cipher.AddressFromPubKey(genPublic)
	_genTime             uint64 = 1000
	_genCoins            uint64 = 1000e6
	_genCoinHours        uint64 = 1000 * 1000
)

func tNow() uint64 {
	return uint64(time.Now().UTC().Unix())
}

func feeCalc(t *Transaction) (uint64, error) {
	return 0, nil
}

func badFeeCalc(t *Transaction) (uint64, error) {
	return 0, errors.New("Bad")
}

func makeNewBlock(uxHash cipher.SHA256) (*Block, error) {
	body := BlockBody{
		Transactions: Transactions{Transaction{}},
	}

	prev := Block{
		Body: body,
		Head: BlockHeader{
			Version:  0x02,
			Time:     100,
			BkSeq:    0,
			Fee:      10,
			PrevHash: cipher.SHA256{},
			BodyHash: body.Hash(),
		}}
	return NewBlock(prev, 100+20, uxHash, Transactions{Transaction{}}, feeCalc)
}

func addTransactionToBlock(t *testing.T, b *Block) Transaction {
	tx := makeTransaction(t)
	b.Body.Transactions = append(b.Body.Transactions, tx)
	return tx
}

func TestNewBlock(t *testing.T) {
	// TODO -- update this test for newBlock changes
	prev := Block{Head: BlockHeader{Version: 0x02, Time: 100, BkSeq: 98}}
	uxHash := testutil.RandSHA256(t)
	txns := Transactions{Transaction{}}
	// invalid txn fees panics
	_, err := NewBlock(prev, 133, uxHash, txns, badFeeCalc)
	require.EqualError(t, err, fmt.Sprintf("Invalid transaction fees: Bad"))

	// no txns panics
	_, err = NewBlock(prev, 133, uxHash, nil, feeCalc)
	require.EqualError(t, err, "Refusing to create block with no transactions")

	_, err = NewBlock(prev, 133, uxHash, Transactions{}, feeCalc)
	require.EqualError(t, err, "Refusing to create block with no transactions")

	// valid block is fine
	fee := uint64(121)
	currentTime := uint64(133)
	b, err := NewBlock(prev, currentTime, uxHash, txns, func(t *Transaction) (uint64, error) {
		return fee, nil
	})
	require.NoError(t, err)
	require.Equal(t, b.Body.Transactions, txns)
	require.Equal(t, b.Head.Fee, fee*uint64(len(txns)))
	require.Equal(t, b.Body, BlockBody{Transactions: txns})
	require.Equal(t, b.Head.PrevHash, prev.HashHeader())
	require.Equal(t, b.Head.Time, currentTime)
	require.Equal(t, b.Head.BkSeq, prev.Head.BkSeq+1)
	require.Equal(t, b.Head.UxHash, uxHash)
}

func TestBlockHashHeader(t *testing.T) {
	uxHash := testutil.RandSHA256(t)
	b, err := makeNewBlock(uxHash)
	require.NoError(t, err)
	require.Equal(t, b.HashHeader(), b.Head.Hash())
	require.NotEqual(t, b.HashHeader(), cipher.SHA256{})
}

func TestBlockHashBody(t *testing.T) {
	uxHash := testutil.RandSHA256(t)
	b, err := makeNewBlock(uxHash)
	require.NoError(t, err)
	require.Equal(t, b.HashBody(), b.Body.Hash())
	hb := b.HashBody()
	hashes := b.Body.Transactions.Hashes()
	tx := addTransactionToBlock(t, b)
	require.NotEqual(t, b.HashBody(), hb)
	hashes = append(hashes, tx.Hash())
	require.Equal(t, b.HashBody(), cipher.Merkle(hashes))
	require.Equal(t, b.HashBody(), b.Body.Hash())
}

func TestNewGenesisBlock(t *testing.T) {
	gb, err := NewGenesisBlock(genAddress, _genCoins, _genTime)
	require.NoError(t, err)

	require.Equal(t, cipher.SHA256{}, gb.Head.PrevHash)
	require.Equal(t, _genTime, gb.Head.Time)
	require.Equal(t, uint64(0), gb.Head.BkSeq)
	require.Equal(t, uint32(0), gb.Head.Version)
	require.Equal(t, uint64(0), gb.Head.Fee)
	require.Equal(t, cipher.SHA256{}, gb.Head.UxHash)

	require.Equal(t, 1, len(gb.Body.Transactions))
	tx := gb.Body.Transactions[0]
	require.Len(t, tx.In, 0)
	require.Len(t, tx.Sigs, 0)
	require.Len(t, tx.Out, 1)

	require.Equal(t, genAddress, tx.Out[0].Address)
	require.Equal(t, _genCoins, tx.Out[0].Coins)
	require.Equal(t, _genCoins, tx.Out[0].Hours)
}

func TestCreateUnspent(t *testing.T) {
	tx := Transaction{}
	tx.PushOutput(genAddress, 11e6, 255)
	bh := BlockHeader{
		Time:  tNow(),
		BkSeq: uint64(1),
	}

	tt := []struct {
		name    string
		txIndex int
		err     error
	}{
		{
			"ok",
			0,
			nil,
		},
		{
			"index overflow",
			10,
			errors.New("Transaction out index is overflow"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			uxout, err := CreateUnspent(bh, tx, tc.txIndex)
			require.Equal(t, tc.err, err)
			if err != nil {
				return
			}
			requireUnspent(t, bh, tx, tc.txIndex, uxout)
		})
	}
}

func TestCreateUnspents(t *testing.T) {
	tx := Transaction{}
	tx.PushOutput(genAddress, 11e6, 255)
	bh := BlockHeader{
		Time:  tNow(),
		BkSeq: uint64(1),
	}
	uxouts := CreateUnspents(bh, tx)
	require.Equal(t, len(uxouts), 1)
	requireValidUnspents(t, bh, tx, uxouts)
}

func requireUnspent(t *testing.T, bh BlockHeader, tx Transaction, txIndex int, ux UxOut) {
	require.Equal(t, bh.Time, ux.Head.Time)
	require.Equal(t, bh.BkSeq, ux.Head.BkSeq)
	require.Equal(t, tx.Hash(), ux.Body.SrcTransaction)
	require.Equal(t, tx.Out[txIndex].Address, ux.Body.Address)
	require.Equal(t, tx.Out[txIndex].Coins, ux.Body.Coins)
	require.Equal(t, tx.Out[txIndex].Hours, ux.Body.Hours)
}

func requireValidUnspents(t *testing.T, bh BlockHeader, tx Transaction,
	uxo UxArray) {
	require.Equal(t, len(tx.Out), len(uxo))
	for i, ux := range uxo {
		require.Equal(t, bh.Time, ux.Head.Time)
		require.Equal(t, bh.BkSeq, ux.Head.BkSeq)
		require.Equal(t, tx.Hash(), ux.Body.SrcTransaction)
		require.Equal(t, tx.Out[i].Address, ux.Body.Address)
		require.Equal(t, tx.Out[i].Coins, ux.Body.Coins)
		require.Equal(t, tx.Out[i].Hours, ux.Body.Hours)
	}
}
