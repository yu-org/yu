package txpool

import (
	"github.com/yu-org/yu/types"
	. "github.com/yu-org/yu/yerror"
)

type TxnCheck func(*types.SignedTxn) error

func Check(checks []TxnCheck, stxn *types.SignedTxn) error {
	for _, check := range checks {
		err := check(stxn)
		if err != nil {
			return err
		}
	}
	return nil
}

//func TripodsCheck(land *Land, stxn *SignedTxn) error {
//	return land.RangeList(func(tri Tripod) error {
//		return tri.CheckTxn(stxn)
//	})
//}

func checkPoolLimit(txnsInPool []*types.SignedTxn, poolsize uint64) error {
	if uint64(len(txnsInPool)) >= poolsize {
		return PoolOverflow
	}
	return nil
}

func checkSignature(stxn *types.SignedTxn) error {
	sig := stxn.Signature
	ecall := stxn.Raw.Ecall
	if !stxn.Pubkey.VerifySignature(ecall.Bytes(), sig) {
		return TxnSignatureErr
	}
	return nil
}

func checkTxnSize(txnMaxSize int, stxn *types.SignedTxn) error {
	if stxn.Size() > txnMaxSize {
		return TxnTooLarge
	}
	return nil
}
