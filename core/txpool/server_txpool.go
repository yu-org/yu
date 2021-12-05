package txpool

//
//import (
//	. "github.com/yu-org/yu/common"
//	"github.com/yu-org/yu/config"
//	. "github.com/yu-org/yu/storage/kv"
//	. "github.com/yu-org/yu/txn"
//	. "github.com/yu-org/yu/yerror"
//	"github.com/sirupsen/logrus"
//	"sync"
//	"time"
//)
//
//// This implementation only use for Master-Worker mode.
//type ServerTxPool struct {
//	sync.RWMutex
//
//	poolSize    uint64
//	TxnMaxSize  int
//	pendingTxns []*SignedTxn
//	Txns        []*SignedTxn
//	packagedIdx int
//	db          KV
//
//	// need to sync txns from p2p
//	ToSyncTxnsChan chan Hash
//	// accept the txn-content of txn-hash from p2p
//	WaitSyncTxnsChan chan *SignedTxn
//	// wait sync txns timeout
//	WaitTxnsTimeout time.Duration
//
//	baseChecks   []TxnCheck
//	tripodChecks []TxnCheck
//}
//
//func NewServerTxPool(cfg *config.TxpoolConf) *ServerTxPool {
//	db, err := NewKV(&cfg.DB)
//	if err != nil {
//		logrus.Panicf("load server txpool error: %s", err.Error())
//	}
//	WaitTxnsTimeout := time.Duration(cfg.WaitTxnsTimeout)
//	return &ServerTxPool{
//		poolSize:         cfg.PoolSize,
//		TxnMaxSize:       cfg.TxnMaxSize,
//		Txns:             make([]*SignedTxn, 0),
//		packagedIdx:      0,
//		db:               db,
//		ToSyncTxnsChan:   make(chan Hash, 1024),
//		WaitSyncTxnsChan: make(chan *SignedTxn, 1024),
//		WaitTxnsTimeout:  WaitTxnsTimeout,
//		baseChecks:       make([]TxnCheck, 0),
//		tripodChecks:     make([]TxnCheck, 0),
//	}
//}
//
//func ServerWithDefaultChecks(cfg *config.TxpoolConf) *ServerTxPool {
//	tp := NewServerTxPool(cfg)
//	return tp.withDefaultBaseChecks()
//}
//
//func (tp *ServerTxPool) withDefaultBaseChecks() *ServerTxPool {
//	tp.baseChecks = []TxnCheck{
//		tp.checkPoolLimit,
//		tp.checkTxnSize,
//		tp.checkSignature,
//	}
//	return tp
//}
//
//func (tp *ServerTxPool) NewEmptySignedTxn() *SignedTxn {
//	return &SignedTxn{}
//}
//
//func (tp *ServerTxPool) NewEmptySignedTxns() SignedTxns {
//	return make([]*SignedTxn, 0)
//}
//
//func (tp *ServerTxPool) PoolSize() uint64 {
//	return tp.poolSize
//}
//
//func (tp *ServerTxPool) WithBaseChecks(checkFns []TxnCheck) ItxPool {
//	tp.baseChecks = append(tp.baseChecks, checkFns...)
//	return tp
//}
//
//func (tp *ServerTxPool) WithTripodChecks(checkFns []TxnCheck) ItxPool {
//	tp.tripodChecks = append(tp.tripodChecks, checkFns...)
//	return tp
//}
//
//// insert into txpool
//func (tp *ServerTxPool) Insert(workerName string, stxn *SignedTxn) (err error) {
//	tp.pendingTxns = append(tp.pendingTxns, stxn)
//	return
//}
//
//// batch insert into txpool
//func (tp *ServerTxPool) BatchInsert(workerName string, txns SignedTxns) error {
//	for _, txn := range txns {
//		err := tp.Insert(workerName, txn)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// package some txns to send to tripods
//func (tp *ServerTxPool) Package(workerName string, numLimit uint64) ([]*SignedTxn, error) {
//	return tp.PackageFor(workerName, numLimit, func(*SignedTxn) error {
//		return nil
//	})
//}
//
//func (tp *ServerTxPool) PackageFor(workerName string, numLimit uint64, filter func(*SignedTxn) error) ([]*SignedTxn, error) {
//	tp.Lock()
//	defer tp.Unlock()
//	stxns := make([]*SignedTxn, 0)
//	for i := 0; i < int(numLimit); i++ {
//		err := filter(tp.Txns[i])
//		if err != nil {
//			return nil, err
//		}
//		stxns = append(stxns, tp.Txns[i])
//		tp.packagedIdx++
//	}
//	return stxns, nil
//}
//
//// get txn content of txn-hash from p2p network
////func (tp *ServerTxPool) SyncTxns(hashes []Hash) error {
////
////	hashesMap := make(map[Hash]bool)
////	tp.RLock()
////	for _, txnHash := range hashes {
////		if !existTxn(txnHash, tp.Txns) {
////			tp.ToSyncTxnsChan <- txnHash
////			hashesMap[txnHash] = true
////		}
////	}
////	tp.RUnlock()
////
////	ticker := time.NewTicker(tp.WaitTxnsTimeout)
////
////	for len(hashesMap) > 0 {
////		select {
////		case stxn := <-tp.WaitSyncTxnsChan:
////			txnHash := stxn.GetRaw().ID()
////			delete(hashesMap, txnHash)
////			err := tp.Insert(workerName, stxn)
////			if err != nil {
////				return err
////			}
////		case <-ticker.C:
////			return WaitTxnsTimeout(hashesMap)
////		}
////	}
////
////	return nil
////}
//
//// remove txns after execute all tripods
//func (tp *ServerTxPool) Flush() error {
//	tp.Lock()
//	tp.Txns = tp.Txns[tp.packagedIdx:]
//	tp.packagedIdx = 0
//	tp.Unlock()
//	return nil
//}
//
//// --------- check txn ------
//
//func (tp *ServerTxPool) BaseCheck(stxn *SignedTxn) error {
//	return Check(tp.baseChecks, stxn)
//}
//
//func (tp *ServerTxPool) TripodsCheck(stxn *SignedTxn) error {
//	return Check(tp.tripodChecks, stxn)
//}
//
//func (tp *ServerTxPool) NecessaryCheck(stxn *SignedTxn) (err error) {
//	err = tp.checkTxnSize(stxn)
//	if err != nil {
//		return
//	}
//	err = tp.checkSignature(stxn)
//	if err != nil {
//		return
//	}
//
//	return tp.TripodsCheck(stxn)
//}
//
//func (tp *ServerTxPool) checkPoolLimit(*SignedTxn) error {
//	return checkPoolLimit(tp.Txns, tp.poolSize)
//}
//
//func (tp *ServerTxPool) checkSignature(stxn *SignedTxn) error {
//	return checkSignature(stxn)
//}
//
//func (tp *ServerTxPool) checkTxnSize(stxn *SignedTxn) error {
//	if stxn.Size() > tp.TxnMaxSize {
//		return TxnTooLarge
//	}
//	return checkTxnSize(tp.TxnMaxSize, stxn)
//}
//
