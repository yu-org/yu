package txpool

import (
	"sync"
	. "yu/txn"
)

type TxPool struct {
	sync.RWMutex

	poolSize   uint64
	timeoutGap uint64
	cache      *TxCache
	Txns       []IsignedTxn
	BaseChecks []BaseCheck
}

func NewTxPool(poolSize uint64) *TxPool {
	return &TxPool{
		poolSize:   poolSize,
		Txns:       make([]IsignedTxn, 0),
		BaseChecks: make([]BaseCheck, 0),
	}
}

func NewWithDefaultChecks(poolSize uint64) *TxPool {
	tp := NewTxPool(poolSize)
	return tp.defaultBaseChecks()
}

func (tp *TxPool) SetCheckFns(checkFns []BaseCheck) *TxPool {
	tp.BaseChecks = checkFns
	return tp
}

func (tp *TxPool) Insert(txn IsignedTxn) (err error) {
	err = tp.baseCheck(txn)
	if err != nil {
		return
	}

	tp.Lock()
	tp.Txns = append(tp.Txns, txn)
	tp.Unlock()
	return
}

func (tp *TxPool) Package(numLimit uint64) ([]IsignedTxn, error) {

}

func (tp *TxPool) Remove() error {

}
