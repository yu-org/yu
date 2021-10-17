package types

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/keypair"
	"github.com/yu-org/yu/utils/codec"
	"strconv"
	"testing"
)

func TestSignedTxns_Remove(t *testing.T) {
	codec.GlobalCodec = &codec.RlpCodec{}

	pubkey, privKey, err := GenKeyPair(Sr25519)
	if err != nil {
		panic("generate key error: " + err.Error())
	}

	var (
		txns   SignedTxns
		txns1  = make(SignedTxns, 3)
		txns2  = make(SignedTxns, 3)
		hashes []Hash
	)
	for i := 0; i < 3; i++ {
		istr := strconv.Itoa(i)
		ecall := &Ecall{
			TripodName: istr,
			ExecName:   istr,
			Params:     JsonString(istr),
		}
		sig, err := privKey.SignData(ecall.Bytes())
		if err != nil {
			t.Fatalf("sign data error: %s", err.Error())
		}
		stxn, err := NewSignedTxn(pubkey.Address(), ecall, pubkey, sig)
		if err != nil {
			t.Fatalf("new SignedTxn error: %s", err.Error())
		}

		txns = append(txns, stxn)

		hashes = append(hashes, stxn.GetTxnHash())
	}

	copy(txns1, txns[:])
	copy(txns2, txns[:])

	removeIdx, restTxns := txns.Remove(hashes[0])
	t.Logf("remove index is %d", removeIdx)
	for _, stxn := range restTxns {
		t.Logf("After removed 0, txn  %s", stxn.GetRaw().Ecall)
	}

	removeIdx, restTxns1 := txns1.Remove(hashes[1])
	t.Logf("remove index is %d", removeIdx)
	for _, stxn := range restTxns1 {
		t.Logf("After removed 1, txn  %s", stxn.GetRaw().Ecall)
	}

	removeIdx, restTxns2 := txns2.Remove(hashes[2])
	t.Logf("remove index is %d", removeIdx)
	for _, stxn := range restTxns2 {
		t.Logf("After removed 2, txn  %s", stxn.GetRaw().Ecall)
	}

}
