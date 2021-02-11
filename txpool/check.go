package txpool

import (
	. "yu/txn"
	. "yu/yerror"
)

type TxnCheck func(txn IunsignedTxn) error

func (tp *TxPool) defaultCheckFns() *TxPool {
	tp.CheckFns = []TxnCheck{
		tp.checkPoolLimit,
	}
	return tp
}

func (tp *TxPool) checkTxn(txn IunsignedTxn) error {
	for _, fn := range tp.CheckFns {
		err := fn(txn)
		if err != nil {
			return err
		}
	}
}

func (tp *TxPool) checkPoolLimit(IunsignedTxn) error {
	if len(tp.Txns) >= tp.poolSize {
		return PoolOverflow
	}
	return nil
}
