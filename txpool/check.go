package txpool

import (
	"yu/tripod"
	. "yu/txn"
	. "yu/yerror"
)

type TxnCheck func(IsignedTxn) error

func (tp *TxPool) withDefaultBaseChecks() *TxPool {
	tp.BaseChecks = []TxnCheck{
		tp.checkExecExist,
		tp.checkPoolLimit,
		tp.checkTxnSize,
		tp.checkDuplicate,
		tp.checkSignature,
	}
	return tp
}

func (tp *TxPool) BaseCheck(stxn IsignedTxn) error {
	for _, check := range tp.BaseChecks {
		err := check(stxn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tp *TxPool) TripodsCheck(stxn IsignedTxn) error {
	return tp.land.RangeList(func(tri tripod.Tripod) error {
		return tri.CheckTxn(stxn)
	})
}

// check if tripod and execution exists
func (tp *TxPool) checkExecExist(stxn IsignedTxn) error {
	ecall := stxn.GetRaw().Ecall()
	tripodName := ecall.TripodName
	execName := ecall.ExecName
	return tp.land.ExistExec(tripodName, execName)
}

func (tp *TxPool) checkPoolLimit(IsignedTxn) error {
	if uint64(len(tp.Txns)) >= tp.poolSize {
		return PoolOverflow
	}
	return nil
}

func (tp *TxPool) checkSignature(stxn IsignedTxn) error {
	sig := stxn.GetSignature()
	hash := stxn.GetTxnHash()
	if !stxn.GetPubkey().VerifySignature(hash.Bytes(), sig) {
		return TxnSignatureErr
	}
	return nil
}

func (tp *TxPool) checkTxnSize(stxn IsignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return nil
}

func (tp *TxPool) checkDuplicate(stxn IsignedTxn) error {
	if existTxn(stxn.GetTxnHash(), tp.Txns) {
		return TxnDuplicate
	}
	return nil
}
