package txpool

import (
	. "github.com/Lawliet-Chan/yu/txn"
	. "github.com/Lawliet-Chan/yu/yerror"
)

type TxnCheck func(*SignedTxn) error

func Check(checks []TxnCheck, stxn *SignedTxn) error {
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

func checkPoolLimit(txnsInPool []*SignedTxn, poolsize uint64) error {
	if uint64(len(txnsInPool)) >= poolsize {
		return PoolOverflow
	}
	return nil
}

func checkSignature(stxn *SignedTxn) error {
	sig := stxn.GetSignature()
	hash := stxn.GetTxnHash()
	if !stxn.GetPubkey().VerifySignature(hash.Bytes(), sig) {
		return TxnSignatureErr
	}
	return nil
}

func checkTxnSize(txnMaxSize int, stxn *SignedTxn) error {
	if stxn.Size() > txnMaxSize {
		return TxnTooLarge
	}
	return nil
}

func checkDuplicate(txnsInPool []*SignedTxn, stxn *SignedTxn) error {
	if existTxn(stxn.GetTxnHash(), txnsInPool) {
		return TxnDuplicate
	}
	return nil
}
