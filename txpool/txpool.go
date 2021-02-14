package txpool

import (
	"sync"
	. "yu/txn"
)

type TxPool struct {
	sync.RWMutex

	poolSize    uint64
	timeoutGap  uint64
	pendingTxns ItxCache
	Txns        []IsignedTxn
	BaseChecks  []BaseCheck
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
	return tp.setDefaultBaseChecks()
}

func (tp *TxPool) SetBaseChecks(checkFns []BaseCheck) *TxPool {
	tp.BaseChecks = checkFns
	return tp
}

// insert into txCache for pending
func (tp *TxPool) Pend(stxn IsignedTxn) (err error) {
	err = tp.baseCheck(stxn)
	if err != nil {
		return
	}

	return tp.pendingTxns.Push(stxn)
}

// insert into txPool for tripods
func (tp *TxPool) Insert(stxn IsignedTxn) (err error) {

}

// package some txns to send to tripods
func (tp *TxPool) Package(numLimit uint64) ([]IsignedTxn, error) {

}

// pop pending txns
func (tp *TxPool) Pop() (IsignedTxn, error) {
	return tp.pendingTxns.Pop()
}

func (tp *TxPool) Remove() error {

}
