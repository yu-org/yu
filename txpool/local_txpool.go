package txpool

import (
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/config"
	. "github.com/Lawliet-Chan/yu/txn"
	ytime "github.com/Lawliet-Chan/yu/utils/time"
	. "github.com/Lawliet-Chan/yu/yerror"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// This implementation only use for Local-Node mode.
type LocalTxPool struct {
	sync.RWMutex

	poolSize   uint64
	TxnMaxSize int

	txnsMap      map[Hash]*SignedTxn
	Txns         SignedTxns
	startPackIdx int

	blockTime uint64
	timeout   time.Duration

	baseChecks   []TxnCheck
	tripodChecks []TxnCheck
}

func NewLocalTxPool(cfg *config.TxpoolConf) *LocalTxPool {
	return &LocalTxPool{
		poolSize:     cfg.PoolSize,
		TxnMaxSize:   cfg.TxnMaxSize,
		txnsMap:      make(map[Hash]*SignedTxn),
		Txns:         make([]*SignedTxn, 0),
		startPackIdx: 0,
		timeout:      time.Duration(cfg.Timeout),
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
func (tp *LocalTxPool) Insert(stxn *SignedTxn) error {
	return tp.BatchInsert(FromArray(stxn))
}

// batch insert into txpool
func (tp *LocalTxPool) BatchInsert(txns SignedTxns) (err error) {
	tp.Lock()
	defer tp.Unlock()
	for _, stxn := range txns {
		if _, ok := tp.txnsMap[stxn.TxnHash]; ok {
			return
		}
		err = tp.BaseCheck(stxn)
		if err != nil {
			return
		}
		err = tp.TripodsCheck(stxn)
		if err != nil {
			return
		}

		tp.Txns = append(tp.Txns, stxn)
		tp.txnsMap[stxn.TxnHash] = stxn
	}
	return
}

// package some txns to send to tripods
func (tp *LocalTxPool) Pack(numLimit uint64) ([]*SignedTxn, error) {
	return tp.PackFor(numLimit, func(*SignedTxn) error {
		return nil
	})
}

func (tp *LocalTxPool) PackFor(numLimit uint64, filter func(*SignedTxn) error) ([]*SignedTxn, error) {
	tp.Lock()
	defer tp.Unlock()
	stxns := make([]*SignedTxn, 0)
	for i := 0; i < int(numLimit); i++ {
		if i >= len(tp.Txns) {
			break
		}
		logrus.Info("********************** pack txn: ", tp.Txns[i].GetTxnHash().String())
		err := filter(tp.Txns[i])
		if err != nil {
			return nil, err
		}
		stxns = append(stxns, tp.Txns[i])
		tp.startPackIdx++
	}
	return stxns, nil
}

func (tp *LocalTxPool) GetTxn(hash Hash) (*SignedTxn, error) {
	tp.RLock()
	defer tp.RUnlock()
	return tp.txnsMap[hash], nil
}

func (tp *LocalTxPool) RemoveTxns(hashes []Hash) error {
	tp.Lock()
	for _, hash := range hashes {
		var idx int
		idx, tp.Txns = tp.Txns.Remove(hash)
		if idx == -1 {
			continue
		}
		delete(tp.txnsMap, hash)
		if idx < tp.startPackIdx {
			tp.startPackIdx--
		}
	}
	tp.Unlock()
	return nil
}

// remove txns after execute all tripods
func (tp *LocalTxPool) Flush() error {
	tp.Lock()
	for _, stxn := range tp.Txns[:tp.startPackIdx] {
		delete(tp.txnsMap, stxn.GetTxnHash())
	}
	tp.Txns = tp.Txns[tp.startPackIdx:]
	tp.startPackIdx = 0
	tp.Unlock()
	return nil
}

func (tp *LocalTxPool) Reset() {
	tp.blockTime = ytime.NowNanoTsU64()
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
