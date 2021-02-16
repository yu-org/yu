package txpool

import (
	"sync"
	. "yu/common"
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
func (tp *TxPool) Insert(num BlockNum, stxn IsignedTxn) (err error) {

}

// package some txns to send to tripods
func (tp *TxPool) Package(numLimit uint64) ([]IsignedTxn, error) {

}

// get txn content of txn-hash from p2p network
func (tp *TxPool) SyncTxns() error {

}

// broadcast txns to p2p network
func (tp *TxPool) BroadcastTxns() error {

}

// pop pending txns
func (tp *TxPool) Pop() (IsignedTxn, error) {
	return tp.pendingTxns.Pop()
}

// remove txns after execute all tripods
func (tp *TxPool) Remove() error {

}
