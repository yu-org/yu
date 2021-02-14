package txpool

import (
	"sync"
	. "yu/txn"
)

type TxCache struct {
	sync.RWMutex
	queueA []IsignedTxn
	queueB []IsignedTxn
}

func NewTxCache() *TxCache {
	return &TxCache{
		queueA: make([]IsignedTxn, 0),
		queueB: make([]IsignedTxn, 0),
	}
}

func (tc *TxCache) Push(stxn IsignedTxn) error {

}

func (tc *TxCache) Pop() (IsignedTxn, error) {

}
