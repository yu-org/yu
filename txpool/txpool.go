package txpool

import (
	"sync"
	"time"
	. "yu/common"
	"yu/config"
	. "yu/txn"
	. "yu/yerror"
)

type TxPool struct {
	sync.RWMutex

	poolSize    uint64
	pendingTxns ItxCache
	Txns        []IsignedTxn
	BaseChecks  []BaseCheck

	// broadcast txns channel
	BcTxnsChan chan IsignedTxn
	// need to sync txns from p2p
	ToSyncTxnsChan chan Hash
	// accept the txn-content of txn-hash from p2p
	WaitSyncTxnsChan chan IsignedTxn
	WaitTxnsTimeout  time.Duration
}

func NewTxPool(cfg *config.TxpoolConf) *TxPool {
	WaitTxnsTimeout := time.Duration(cfg.WaitTxnsTimeout)
	return &TxPool{
		poolSize:         cfg.PoolSize,
		Txns:             make([]IsignedTxn, 0),
		BaseChecks:       make([]BaseCheck, 0),
		BcTxnsChan:       make(chan IsignedTxn, 1024),
		ToSyncTxnsChan:   make(chan Hash, 1024),
		WaitSyncTxnsChan: make(chan IsignedTxn, 1024),
		WaitTxnsTimeout:  WaitTxnsTimeout,
	}
}

func NewWithDefaultChecks(cfg *config.TxpoolConf) *TxPool {
	tp := NewTxPool(cfg)
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
func (tp *TxPool) SyncTxns(txnHashes []Hash) error {
	hashes := make(map[Hash]bool)
	for _, txnHash := range txnHashes {
		tp.ToSyncTxnsChan <- txnHash
		hashes[txnHash] = true
	}

	ticker := time.NewTicker(tp.WaitTxnsTimeout)

	for len(hashes) > 0 {
		select {
		case stxn := <-tp.WaitSyncTxnsChan:
			txnHash := stxn.GetRaw().ID()
			delete(hashes, txnHash)
			err := tp.Pend(stxn)
			if err != nil {
				return err
			}
		case <-ticker.C:
			return WaitTxnsTimeout(hashes)
		}
	}

	return nil
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
