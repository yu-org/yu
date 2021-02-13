package txpool

import (
	. "yu/txn"
	. "yu/yerror"
)

type BaseCheck func(txn IsignedTxn) error

func (tp *TxPool) defaultBaseChecks() *TxPool {
	tp.CheckFns = []BaseCheck{
		tp.checkPoolLimit,
	}
	return tp
}

func (tp *TxPool) baseCheck(txn IsignedTxn) error {
	for _, fn := range tp.CheckFns {
		err := fn(txn)
		if err != nil {
			return err
		}
	}
}

func (tp *TxPool) checkPoolLimit(IsignedTxn) error {
	if uint64(len(tp.Txns)) >= tp.poolSize {
		return PoolOverflow
	}
	return nil
}
