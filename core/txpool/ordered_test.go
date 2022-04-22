package txpool

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/core/types"
	"testing"
)

var (
	tx1, tx2, tx3    *SignedTxn
	caller1, caller2 Address
)

func init() {
	pubkey1, privkey1 := keypair.GenSrKey([]byte("yu"))
	caller1 = pubkey1.Address()
	pubkey2, privkey2 := keypair.GenSrKey([]byte("boyi"))
	caller2 = pubkey2.Address()

	ecall1 := &Ecall{LeiPrice: 10}
	hash1, err := ecall1.Hash()
	if err != nil {
		panic(err)
	}
	sig1, err := privkey1.SignData(hash1)
	if err != nil {
		panic(err)
	}
	ecall2 := &Ecall{LeiPrice: 30}
	hash2, err := ecall2.Hash()
	if err != nil {
		panic(err)
	}
	sig2, err := privkey1.SignData(hash2)
	if err != nil {
		panic(err)
	}
	ecall3 := &Ecall{LeiPrice: 20}
	hash3, err := ecall3.Hash()
	if err != nil {
		panic(err)
	}
	sig3, err := privkey2.SignData(hash3)
	if err != nil {
		panic(err)
	}

	tx1, err = NewSignedTxn(caller1, ecall1, pubkey1, sig1)
	if err != nil {
		panic(err)
	}
	tx2, err = NewSignedTxn(caller1, ecall2, pubkey1, sig2)
	if err != nil {
		panic(err)
	}
	tx3, err = NewSignedTxn(caller2, ecall3, pubkey1, sig3)
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
