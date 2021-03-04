package txpool

import (
	. "yu/tripod"
	. "yu/txn"
	. "yu/yerror"
)

type TxnCheck func(IsignedTxn) error

func BaseCheck(checks []TxnCheck, stxn IsignedTxn) error {
	for _, check := range checks {
		err := check(stxn)
		if err != nil {
			return err
		}
	}
	return nil
}

func TripodsCheck(land *Land, stxn IsignedTxn) error {
	return land.RangeList(func(tri Tripod) error {
		return tri.CheckTxn(stxn)
	})
}

// check if tripod and execution exists
func checkExecExist(land *Land, stxn IsignedTxn) error {
	ecall := stxn.GetRaw().Ecall()
	tripodName := ecall.TripodName
	execName := ecall.ExecName
	return land.ExistExec(tripodName, execName)
}

func checkPoolLimit(txnsInPool []IsignedTxn, poolsize uint64) error {
	if uint64(len(txnsInPool)) >= poolsize {
		return PoolOverflow
	}
	return nil
}

func checkSignature(stxn IsignedTxn) error {
	sig := stxn.GetSignature()
	hash := stxn.GetTxnHash()
	if !stxn.GetPubkey().VerifySignature(hash.Bytes(), sig) {
		return TxnSignatureErr
	}
	return nil
}

func checkTxnSize(txnMaxSize int, stxn IsignedTxn) error {
	if stxn.Size() > txnMaxSize {
		return TxnTooLarge
	}
	return nil
}

func checkDuplicate(txnsInPool []IsignedTxn, stxn IsignedTxn) error {
	if existTxn(stxn.GetTxnHash(), txnsInPool) {
		return TxnDuplicate
	}
	return nil
}
