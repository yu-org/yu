package txpool

import (
	. "yu/txn"
	. "yu/yerror"
)

type BaseCheck func(IsignedTxn) error

func (tp *TxPool) setDefaultBaseChecks() *TxPool {
	tp.BaseChecks = []BaseCheck{
		tp.checkPoolLimit,
		tp.checkSignature,
	}
	return tp
}

func (tp *TxPool) baseCheck(stxn IsignedTxn) error {
	for _, fn := range tp.BaseChecks {
		err := fn(stxn)
		if err != nil {
			return err
		}
	}
	return nil
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
