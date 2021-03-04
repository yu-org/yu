package txpool

import (
	"sync"
	"time"
	. "yu/common"
	"yu/config"
	. "yu/storage/kv"
	"yu/tripod"
	. "yu/txn"
	. "yu/yerror"
)

type LocalTxPool struct {
	sync.RWMutex

	poolSize    uint64
	TxnMaxSize  int
	pendingTxns []IsignedTxn
	Txns        []IsignedTxn
	packagedIdx int
	db          KV

	// broadcast txns channel
	BcTxnsChan chan IsignedTxn
	// need to sync txns from p2p
	ToSyncTxnsChan chan Hash
	// accept the txn-content of txn-hash from p2p
	WaitSyncTxnsChan chan IsignedTxn
	// wait sync txns timeout
	WaitTxnsTimeout time.Duration

	BaseChecks []TxnCheck
	land       *tripod.Land
}

func NewLocalTxPool(cfg *config.TxpoolConf, land *tripod.Land) (*LocalTxPool, error) {
	db, err := NewKV(&cfg.DB)
	if err != nil {
		return nil, err
	}
	WaitTxnsTimeout := time.Duration(cfg.WaitTxnsTimeout)
	return &LocalTxPool{
		poolSize:         cfg.PoolSize,
		TxnMaxSize:       cfg.TxnMaxSize,
		Txns:             make([]IsignedTxn, 0),
		packagedIdx:      0,
		db:               db,
		BcTxnsChan:       make(chan IsignedTxn, 1024),
		ToSyncTxnsChan:   make(chan Hash, 1024),
		WaitSyncTxnsChan: make(chan IsignedTxn, 1024),
		WaitTxnsTimeout:  WaitTxnsTimeout,
		BaseChecks:       make([]TxnCheck, 0),
		land:             land,
	}, nil
}

func NewWithDefaultChecks(cfg *config.TxpoolConf, land *tripod.Land) (*LocalTxPool, error) {
	tp, err := NewLocalTxPool(cfg, land)
	if err != nil {
		return nil, err
	}
	return tp.withDefaultBaseChecks(), nil
}

func (tp *LocalTxPool) withDefaultBaseChecks() *LocalTxPool {
	tp.BaseChecks = []TxnCheck{
		tp.checkExecExist,
		tp.checkPoolLimit,
		tp.checkTxnSize,
		tp.checkDuplicate,
		tp.checkSignature,
	}
	return tp
}

func (tp *LocalTxPool) PoolSize() uint64 {
	return tp.poolSize
}

func (tp *LocalTxPool) WithBaseChecks(checkFns []TxnCheck) ItxPool {
	tp.BaseChecks = checkFns
	return tp
}

// insert into txCache for pending
func (tp *LocalTxPool) Insert(stxn IsignedTxn) (err error) {
	err = tp.BaseCheck(stxn)
	if err != nil {
		return
	}
	err = tp.TripodsCheck(stxn)
	if err != nil {
		return
	}

	tp.pendingTxns = append(tp.pendingTxns, stxn)
	return
}

// package some txns to send to tripods
func (tp *LocalTxPool) Package(numLimit uint64) ([]IsignedTxn, error) {
	return tp.PackageFor(numLimit, func(IsignedTxn) error {
		return nil
	})
}

func (tp *LocalTxPool) PackageFor(numLimit uint64, filter func(IsignedTxn) error) ([]IsignedTxn, error) {
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
func (tp *LocalTxPool) SyncTxns(hashes []Hash) error {

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
			err := tp.Insert(stxn)
			if err != nil {
				return err
			}
		case <-ticker.C:
			return WaitTxnsTimeout(hashesMap)
		}
	}

	return nil
}

// broadcast txn to p2p network
func (tp *LocalTxPool) BroadcastTxn(stxn IsignedTxn) {
	tp.BcTxnsChan <- stxn
}

// remove txns after execute all tripods
func (tp *LocalTxPool) Remove() error {
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

// --------- check txn ------

func (tp *LocalTxPool) BaseCheck(stxn IsignedTxn) error {
	return BaseCheck(tp.BaseChecks, stxn)
}

func (tp *LocalTxPool) TripodsCheck(stxn IsignedTxn) error {
	return TripodsCheck(tp.land, stxn)
}

// check if tripod and execution exists
func (tp *LocalTxPool) checkExecExist(stxn IsignedTxn) error {
	return checkExecExist(tp.land, stxn)
}

func (tp *LocalTxPool) checkPoolLimit(IsignedTxn) error {
	return checkPoolLimit(tp.Txns, tp.poolSize)
}

func (tp *LocalTxPool) checkSignature(stxn IsignedTxn) error {
	return checkSignature(stxn)
}

func (tp *LocalTxPool) checkTxnSize(stxn IsignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return checkTxnSize(tp.TxnMaxSize, stxn)
}

func (tp *LocalTxPool) checkDuplicate(stxn IsignedTxn) error {
	return checkDuplicate(tp.Txns, stxn)
}
