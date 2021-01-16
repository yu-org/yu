package txn

import (
	"github.com/pkg/errors"
	"sync"
)

var PoolOverflowErr error = errors.New("pool size is full")
var TxnSignatureErr error = errors.New("the signature of Txn illegal")

type TxPool struct {
	sync.RWMutex

	poolSize   int
	SignedTxns []Itxn
}

func NewTxPool(poolSize int) *TxPool {
	return &TxPool{
		poolSize:   poolSize,
		SignedTxns: make([]Itxn, 0),
	}
}

func (tp *TxPool) InsertTxn(txn Itxn) (err error) {
	err = tp.checkPoolLimit()
	if err != nil {
		return
	}

	err = txn.Verify()
	if err != nil {
		return
	}

	tp.Lock()
	tp.SignedTxns = append(tp.SignedTxns, txn)
	tp.Unlock()
	return
}

func (tp *TxPool) checkPoolLimit() error {
	if len(tp.SignedTxns) >= tp.poolSize {
		return PoolOverflowErr
	}
	return nil
}
