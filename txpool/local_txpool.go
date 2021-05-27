package txpool

import (
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/config"
	. "github.com/Lawliet-Chan/yu/txn"
	. "github.com/Lawliet-Chan/yu/yerror"
	"sync"
)

// This implementation only use for Local-Node mode.
type LocalTxPool struct {
	sync.RWMutex

	poolSize    uint64
	TxnMaxSize  int
	pendingTxns []*SignedTxn
	Txns        []*SignedTxn
	packagedIdx int

	// need to sync txns from p2p
	//ToSyncTxnsChan chan []Hash
	// accept the txn-content of txn-hash from p2p
	//WaitSyncTxnsChan chan *SignedTxn
	// wait sync txns timeout
	//WaitTxnsTimeout time.Duration

	baseChecks   []TxnCheck
	tripodChecks []TxnCheck
}

func NewLocalTxPool(cfg *config.TxpoolConf) *LocalTxPool {
	return &LocalTxPool{
		poolSize:    cfg.PoolSize,
		TxnMaxSize:  cfg.TxnMaxSize,
		Txns:        make([]*SignedTxn, 0),
		packagedIdx: 0,
		//ToSyncTxnsChan:   make(chan []Hash),
		//WaitSyncTxnsChan: make(chan *SignedTxn, 1024),
		//WaitTxnsTimeout:  WaitTxnsTimeout,
		baseChecks:   make([]TxnCheck, 0),
		tripodChecks: make([]TxnCheck, 0),
	}
}

func LocalWithDefaultChecks(cfg *config.TxpoolConf) *LocalTxPool {
	tp := NewLocalTxPool(cfg)
	return tp.withDefaultBaseChecks()
}

func (tp *LocalTxPool) withDefaultBaseChecks() *LocalTxPool {
	tp.baseChecks = []TxnCheck{
		tp.checkPoolLimit,
		tp.checkTxnSize,
		tp.checkDuplicate,
		tp.checkSignature,
	}
	return tp
}

func (tp *LocalTxPool) NewEmptySignedTxn() *SignedTxn {
	return &SignedTxn{}
}

func (tp *LocalTxPool) NewEmptySignedTxns() SignedTxns {
	return make([]*SignedTxn, 0)
}

func (tp *LocalTxPool) PoolSize() uint64 {
	return tp.poolSize
}

func (tp *LocalTxPool) WithBaseChecks(checkFns []TxnCheck) ItxPool {
	tp.baseChecks = append(tp.baseChecks, checkFns...)
	return tp
}

func (tp *LocalTxPool) WithTripodChecks(checkFns []TxnCheck) ItxPool {
	tp.tripodChecks = append(tp.tripodChecks, checkFns...)
	return tp
}

// insert into txpool
func (tp *LocalTxPool) Insert(_ string, stxn *SignedTxn) (err error) {
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

// batch insert into txpool
func (tp *LocalTxPool) BatchInsert(_ string, txns SignedTxns) error {
	for _, txn := range txns {
		err := tp.Insert("", txn)
		if err != nil {
			return err
		}
	}
	return nil
}

// package some txns to send to tripods
func (tp *LocalTxPool) Package(_ string, numLimit uint64) ([]*SignedTxn, error) {
	return tp.PackageFor("", numLimit, func(*SignedTxn) error {
		return nil
	})
}

func (tp *LocalTxPool) PackageFor(_ string, numLimit uint64, filter func(*SignedTxn) error) ([]*SignedTxn, error) {
	tp.Lock()
	defer tp.Unlock()
	stxns := make([]*SignedTxn, 0)
	for i := 0; i < int(numLimit); i++ {
		if i >= len(tp.Txns) {
			break
		}
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
//func (tp *LocalTxPool) SyncTxns(hashes []Hash) error {
//
//	hashesMap := make(map[Hash]bool)
//	toSyncHashes := make([]Hash, 0)
//
//	tp.RLock()
//	for _, txnHash := range hashes {
//		if !existTxn(txnHash, tp.Txns) {
//			toSyncHashes = append(toSyncHashes, txnHash)
//			hashesMap[txnHash] = true
//		}
//	}
//	tp.RUnlock()
//
//	tp.ToSyncTxnsChan <- toSyncHashes
//
//	ticker := time.NewTicker(tp.WaitTxnsTimeout)
//
//	for len(hashesMap) > 0 {
//		select {
//		case stxn := <-tp.WaitSyncTxnsChan:
//			txnHash := stxn.GetTxnHash()
//			delete(hashesMap, txnHash)
//			err := tp.Insert("", stxn)
//			if err != nil {
//				return err
//			}
//		case <-ticker.C:
//			return WaitTxnsTimeout(hashesMap)
//		}
//	}
//
//	return nil
//}

// remove txns after execute all tripods
func (tp *LocalTxPool) Flush() error {
	tp.Lock()
	tp.Txns = tp.Txns[tp.packagedIdx:]
	tp.packagedIdx = 0
	tp.Unlock()
	return nil
}

func existTxn(hash Hash, txns []*SignedTxn) bool {
	for _, txn := range txns {
		if txn.GetTxnHash() == hash {
			return true
		}
	}
	return false
}

// --------- check txn ------

func (tp *LocalTxPool) BaseCheck(stxn *SignedTxn) error {
	return Check(tp.baseChecks, stxn)
}

func (tp *LocalTxPool) TripodsCheck(stxn *SignedTxn) error {
	return Check(tp.tripodChecks, stxn)
}

func (tp *LocalTxPool) NecessaryCheck(stxn *SignedTxn) (err error) {
	err = tp.checkTxnSize(stxn)
	if err != nil {
		return
	}
	err = tp.checkSignature(stxn)
	if err != nil {
		return
	}

	return tp.TripodsCheck(stxn)
}

func (tp *LocalTxPool) checkPoolLimit(*SignedTxn) error {
	return checkPoolLimit(tp.Txns, tp.poolSize)
}

func (tp *LocalTxPool) checkSignature(stxn *SignedTxn) error {
	return checkSignature(stxn)
}

func (tp *LocalTxPool) checkTxnSize(stxn *SignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return checkTxnSize(tp.TxnMaxSize, stxn)
}

func (tp *LocalTxPool) checkDuplicate(stxn *SignedTxn) error {
	return checkDuplicate(tp.Txns, stxn)
}
