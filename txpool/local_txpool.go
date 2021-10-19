package txpool

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/types"
	. "github.com/yu-org/yu/yerror"
	"sync"
	"time"
)

// This implementation only use for Local-Node mode.
type LocalTxPool struct {
	sync.RWMutex

	poolSize   uint64
	TxnMaxSize int

	txnsMap      map[Hash]*types.SignedTxn
	Txns         types.SignedTxns
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
		txnsMap:      make(map[Hash]*types.SignedTxn),
		Txns:         make([]*types.SignedTxn, 0),
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

func (tp *LocalTxPool) NewEmptySignedTxn() *types.SignedTxn {
	return &types.SignedTxn{}
}

func (tp *LocalTxPool) NewEmptySignedTxns() types.SignedTxns {
	return make([]*types.SignedTxn, 0)
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
func (tp *LocalTxPool) Insert(stxn *types.SignedTxn) error {
	return tp.BatchInsert(types.FromArray(stxn))
}

// batch insert into txpool
func (tp *LocalTxPool) BatchInsert(txns types.SignedTxns) (err error) {
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
func (tp *LocalTxPool) Pack(numLimit uint64) ([]*types.SignedTxn, error) {
	return tp.PackFor(numLimit, func(*types.SignedTxn) error {
		return nil
	})
}

func (tp *LocalTxPool) PackFor(numLimit uint64, filter func(*types.SignedTxn) error) ([]*types.SignedTxn, error) {
	tp.Lock()
	defer tp.Unlock()
	stxns := make([]*types.SignedTxn, 0)
	for i := 0; i < int(numLimit); i++ {
		if i >= len(tp.Txns) {
			break
		}
		logrus.Info("********************** pack txn: ", tp.Txns[i].TxnHash.String())
		err := filter(tp.Txns[i])
		if err != nil {
			return nil, err
		}
		stxns = append(stxns, tp.Txns[i])
		tp.startPackIdx++
	}
	return stxns, nil
}

func (tp *LocalTxPool) GetTxn(hash Hash) (*types.SignedTxn, error) {
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
func (tp *LocalTxPool) Reset() error {
	tp.Lock()
	for _, stxn := range tp.Txns[:tp.startPackIdx] {
		delete(tp.txnsMap, stxn.TxnHash)
	}
	tp.Txns = tp.Txns[tp.startPackIdx:]
	tp.startPackIdx = 0
	tp.Unlock()
	return nil
}

// --------- check txn ------

func (tp *LocalTxPool) BaseCheck(stxn *types.SignedTxn) error {
	return Check(tp.baseChecks, stxn)
}

func (tp *LocalTxPool) TripodsCheck(stxn *types.SignedTxn) error {
	return Check(tp.tripodChecks, stxn)
}

func (tp *LocalTxPool) NecessaryCheck(stxn *types.SignedTxn) (err error) {
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

func (tp *LocalTxPool) checkPoolLimit(*types.SignedTxn) error {
	return checkPoolLimit(tp.Txns, tp.poolSize)
}

func (tp *LocalTxPool) checkSignature(stxn *types.SignedTxn) error {
	return checkSignature(stxn)
}

func (tp *LocalTxPool) checkTxnSize(stxn *types.SignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return checkTxnSize(tp.TxnMaxSize, stxn)
}
