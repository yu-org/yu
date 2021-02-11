package txpool

import (
	"sync"
	. "yu/txn"
)

type TxPool struct {
	sync.RWMutex

	poolSize int
	Txns     []IunsignedTxn
	CheckFns []TxnCheck
}

func NewTxPool(poolSize int) *TxPool {
	return &TxPool{
		poolSize: poolSize,
		Txns:     make([]IunsignedTxn, 0),
	}
}

func NewWithDefaultChecks(poolSize int) *TxPool {
	tp := NewTxPool(poolSize)
	return tp.defaultCheckFns()
}

func (tp *TxPool) SetCheckFns(checkFns []TxnCheck) *TxPool {
	tp.CheckFns = checkFns
	return tp
}

func (tp *TxPool) Insert(txn IunsignedTxn) (err error) {
	err = tp.checkTxn(txn)
	if err != nil {
		return
	}

	tp.Lock()
	tp.Txns = append(tp.Txns, txn)
	tp.Unlock()
	return
}

func (tp *TxPool) Remove() error {

}
