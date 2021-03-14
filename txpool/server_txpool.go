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

type ServerTxPool struct {
	sync.RWMutex

	poolSize    uint64
	TxnMaxSize  int
	pendingTxns []IsignedTxn
	Txns        []IsignedTxn
	packagedIdx int
	db          KV

	// need to sync txns from p2p
	ToSyncTxnsChan chan Hash
	// accept the txn-content of txn-hash from p2p
	WaitSyncTxnsChan chan IsignedTxn
	// wait sync txns timeout
	WaitTxnsTimeout time.Duration

	BaseChecks []TxnCheck
	land       *tripod.Land
}

func NewServerTxPool(cfg *config.TxpoolConf, land *tripod.Land) (*ServerTxPool, error) {
	db, err := NewKV(&cfg.DB)
	if err != nil {
		return nil, err
	}
	WaitTxnsTimeout := time.Duration(cfg.WaitTxnsTimeout)
	return &ServerTxPool{
		poolSize:         cfg.PoolSize,
		TxnMaxSize:       cfg.TxnMaxSize,
		Txns:             make([]IsignedTxn, 0),
		packagedIdx:      0,
		db:               db,
		ToSyncTxnsChan:   make(chan Hash, 1024),
		WaitSyncTxnsChan: make(chan IsignedTxn, 1024),
		WaitTxnsTimeout:  WaitTxnsTimeout,
		BaseChecks:       make([]TxnCheck, 0),
		land:             land,
	}, nil
}

func ServerWithDefaultChecks(cfg *config.TxpoolConf, land *tripod.Land) (*ServerTxPool, error) {
	tp, err := NewServerTxPool(cfg, land)
	if err != nil {
		return nil, err
	}
	return tp.withDefaultBaseChecks(), nil
}

func (tp *ServerTxPool) withDefaultBaseChecks() *ServerTxPool {
	tp.BaseChecks = []TxnCheck{
		tp.checkExecExist,
		tp.checkPoolLimit,
		tp.checkTxnSize,
		tp.checkDuplicate,
		tp.checkSignature,
	}
	return tp
}

func (tp *ServerTxPool) PoolSize() uint64 {
	return tp.poolSize
}

func (tp *ServerTxPool) WithBaseChecks(checkFns []TxnCheck) ItxPool {
	tp.BaseChecks = checkFns
	return tp
}

// insert into txpool
func (tp *ServerTxPool) Insert(workerIP string, stxn IsignedTxn) (err error) {
	tp.pendingTxns = append(tp.pendingTxns, stxn)
	return
}

// batch insert into txpool
func (tp *ServerTxPool) BatchInsert(workerIP string, txns SignedTxns) error {
	for _, txn := range txns {
		err := tp.Insert(workerIP, txn)
		if err != nil {
			return err
		}
	}
	return nil
}

// package some txns to send to tripods
func (tp *ServerTxPool) Package(numLimit uint64) ([]IsignedTxn, error) {
	return tp.PackageFor(numLimit, func(IsignedTxn) error {
		return nil
	})
}

func (tp *ServerTxPool) PackageFor(numLimit uint64, filter func(IsignedTxn) error) ([]IsignedTxn, error) {
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
func (tp *ServerTxPool) SyncTxns(hashes []Hash) error {

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

// remove txns after execute all tripods
func (tp *ServerTxPool) Remove() error {
	tp.Lock()
	tp.Txns = tp.Txns[tp.packagedIdx:]
	tp.packagedIdx = 0
	tp.Unlock()
	return nil
}

// --------- check txn ------

func (tp *ServerTxPool) BaseCheck(stxn IsignedTxn) error {
	return BaseCheck(tp.BaseChecks, stxn)
}

func (tp *ServerTxPool) TripodsCheck(stxn IsignedTxn) error {
	return TripodsCheck(tp.land, stxn)
}

// check if tripod and execution exists
func (tp *ServerTxPool) checkExecExist(stxn IsignedTxn) error {
	return checkExecExist(tp.land, stxn)
}

func (tp *ServerTxPool) checkPoolLimit(IsignedTxn) error {
	return checkPoolLimit(tp.Txns, tp.poolSize)
}

func (tp *ServerTxPool) checkSignature(stxn IsignedTxn) error {
	return checkSignature(stxn)
}

func (tp *ServerTxPool) checkTxnSize(stxn IsignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return checkTxnSize(tp.TxnMaxSize, stxn)
}

func (tp *ServerTxPool) checkDuplicate(stxn IsignedTxn) error {
	return checkDuplicate(tp.Txns, stxn)
}
