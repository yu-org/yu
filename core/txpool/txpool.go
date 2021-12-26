package txpool

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	"sync"
	"time"
)

// This implementation only use for Local-Node mode.
type TxPool struct {
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

func NewTxPool(cfg *config.TxpoolConf) *TxPool {
	return &TxPool{
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

func LocalWithDefaultChecks(cfg *config.TxpoolConf) *TxPool {
	tp := NewTxPool(cfg)
	return tp.withDefaultBaseChecks()
}

func (tp *TxPool) withDefaultBaseChecks() *TxPool {
	tp.baseChecks = []TxnCheck{
		tp.checkPoolLimit,
		tp.checkTxnSize,
		tp.checkSignature,
	}
	return tp
}

func (tp *TxPool) NewEmptySignedTxn() *SignedTxn {
	return &SignedTxn{}
}

func (tp *TxPool) NewEmptySignedTxns() SignedTxns {
	return make([]*SignedTxn, 0)
}

func (tp *TxPool) PoolSize() uint64 {
	return tp.poolSize
}

func (tp *TxPool) WithBaseChecks(checkFns []TxnCheck) ItxPool {
	tp.baseChecks = append(tp.baseChecks, checkFns...)
	return tp
}

func (tp *TxPool) WithTripodChecks(checkFns []TxnCheck) ItxPool {
	tp.tripodChecks = append(tp.tripodChecks, checkFns...)
	return tp
}

// insert into txpool
func (tp *TxPool) Insert(stxn *SignedTxn) error {
	return tp.BatchInsert(FromArray(stxn))
}

// batch insert into txpool
func (tp *TxPool) BatchInsert(txns SignedTxns) (err error) {
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
func (tp *TxPool) Pack(numLimit uint64) ([]*SignedTxn, error) {
	return tp.PackFor(numLimit, func(*SignedTxn) error {
		return nil
	})
}

func (tp *TxPool) PackFor(numLimit uint64, filter func(*SignedTxn) error) ([]*SignedTxn, error) {
	tp.Lock()
	defer tp.Unlock()
	stxns := make([]*SignedTxn, 0)
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

func (tp *TxPool) GetTxn(hash Hash) (*SignedTxn, error) {
	tp.RLock()
	defer tp.RUnlock()
	return tp.txnsMap[hash], nil
}

func (tp *TxPool) RemoveTxns(hashes []Hash) error {
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
func (tp *TxPool) Reset() error {
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

func (tp *TxPool) BaseCheck(stxn *SignedTxn) error {
	return Check(tp.baseChecks, stxn)
}

func (tp *TxPool) TripodsCheck(stxn *SignedTxn) error {
	return Check(tp.tripodChecks, stxn)
}

func (tp *TxPool) NecessaryCheck(stxn *SignedTxn) (err error) {
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

func (tp *TxPool) checkPoolLimit(*SignedTxn) error {
	return checkPoolLimit(tp.Txns, tp.poolSize)
}

func (tp *TxPool) checkSignature(stxn *SignedTxn) error {
	return checkSignature(stxn)
}

func (tp *TxPool) checkTxnSize(stxn *SignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return checkTxnSize(tp.TxnMaxSize, stxn)
}
