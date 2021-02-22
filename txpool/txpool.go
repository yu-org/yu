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
	// the last block height
	height      BlockNum
	poolSize    uint64
	TxnMaxSize  int
	pendingTxns ItxCache
	Txns        []IsignedTxn
	packagedIdx int

	// broadcast txns channel
	BcTxnsChan chan IsignedTxn
	// need to sync txns from p2p
	ToSyncTxnsChan chan Hash
	// accept the txn-content of txn-hash from p2p
	WaitSyncTxnsChan chan IsignedTxn
	// wait sync txns timeout
	WaitTxnsTimeout time.Duration

	BaseChecks []BaseCheck
}

func NewTxPool(cfg *config.TxpoolConf, height BlockNum) *TxPool {
	WaitTxnsTimeout := time.Duration(cfg.WaitTxnsTimeout)
	return &TxPool{
		height:           height,
		poolSize:         cfg.PoolSize,
		TxnMaxSize:       cfg.TxnMaxSize,
		Txns:             make([]IsignedTxn, 0),
		packagedIdx:      0,
		BcTxnsChan:       make(chan IsignedTxn, 1024),
		ToSyncTxnsChan:   make(chan Hash, 1024),
		WaitSyncTxnsChan: make(chan IsignedTxn, 1024),
		WaitTxnsTimeout:  WaitTxnsTimeout,
		BaseChecks:       make([]BaseCheck, 0),
	}
}

func NewWithDefaultChecks(cfg *config.TxpoolConf, height BlockNum) *TxPool {
	tp := NewTxPool(cfg, height)
	return tp.withDefaultBaseChecks()
}

func (tp *TxPool) PoolSize() uint64 {
	return tp.poolSize
}

func (tp *TxPool) WithBaseChecks(checkFns []BaseCheck) ItxPool {
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
func (tp *TxPool) Insert(height BlockNum, stxn IsignedTxn) (err error) {

}

// package some txns to send to tripods
func (tp *TxPool) Package(numLimit uint64) ([]IsignedTxn, error) {
	return tp.PackageFor(numLimit, func(IsignedTxn) error {
		return nil
	})
}

func (tp *TxPool) PackageFor(numLimit uint64, filter func(IsignedTxn) error) ([]IsignedTxn, error) {
	tp.Lock()
	defer tp.Unlock()
	stxns := make([]IsignedTxn, 0)
	for i := 0; i < int(numLimit); i++ {
		err := filter(tp.Txns[i])
		if err != nil {
			return nil, err
		}
		stxns = append(stxns, tp.Txns[i])
		tp.packagedIdx++
	}
	return stxns, nil
}

// get txn content of txn-hash from p2p network
func (tp *TxPool) SyncTxns(hashes []Hash) error {

	hashesMap := make(map[Hash]bool)
	tp.RLock()
	for _, txnHash := range hashes {
		if !existTxn(txnHash, tp.Txns) {
			tp.ToSyncTxnsChan <- txnHash
			hashesMap[txnHash] = true
		}
	}
	tp.RUnlock()

	ticker := time.NewTicker(tp.WaitTxnsTimeout)

	for len(hashesMap) > 0 {
		select {
		case stxn := <-tp.WaitSyncTxnsChan:
			txnHash := stxn.GetRaw().ID()
			delete(hashesMap, txnHash)
			err := tp.Pend(stxn)
			if err != nil {
				return err
			}
		case <-ticker.C:
			return WaitTxnsTimeout(hashesMap)
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
	tp.Lock()
	tp.Txns = tp.Txns[tp.packagedIdx:]
	tp.packagedIdx = 0
	tp.Unlock()
	return nil
}

func existTxn(hash Hash, txns []IsignedTxn) bool {
	for _, txn := range txns {
		if txn.GetTxnHash() == hash {
			return true
		}
	}
	return false
}
