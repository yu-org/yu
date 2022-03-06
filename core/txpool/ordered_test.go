package txpool

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/core/types"
	"testing"
)

var (
	tx1, tx2, tx3 *SignedTxn
)

func init() {
	pubkey, privkey := keypair.GenSrKey([]byte("yu"))
	caller := pubkey.Address()

	ecall1 := &Ecall{LeiPrice: 10}
	sig1, err := privkey.SignData(ecall1.Bytes())
	if err != nil {
		panic(err)
	}

	ecall2 := &Ecall{LeiPrice: 30}
	sig2, err := privkey.SignData(ecall2.Bytes())
	if err != nil {
		panic(err)
	}

	ecall3 := &Ecall{LeiPrice: 20}
	sig3, err := privkey.SignData(ecall3.Bytes())
	if err != nil {
		panic(err)
	}

	tx1, err = NewSignedTxn(caller, ecall1, pubkey, sig1)
	if err != nil {
		panic(err)
	}
	tx2, err = NewSignedTxn(caller, ecall2, pubkey, sig2)
	if err != nil {
		panic(err)
	}
	tx3, err = NewSignedTxn(caller, ecall3, pubkey, sig3)
	if err != nil {
		panic(err)
	}
}

func TestOrdered(t *testing.T) {
	correctOrder := []*SignedTxn{tx2, tx3, tx1}

	otxns := newOrderedTxns()
	otxns.insert(tx1)
	otxns.insert(tx2)
	otxns.insert(tx3)
	txns := otxns.gets(3, func(txn *SignedTxn) bool {
		return true
	})
	for i, txn := range txns {
		assert.Equal(t, txn, correctOrder[i])
	}
}
