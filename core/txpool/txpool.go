package txpool

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/metrics"
	"sync"
)

type TxPool struct {
	// sync.RWMutex

	nodeType int

	capacity   int
	TxnMaxSize int

	unpackedTxns IunpackedTxns

	topicLock         sync.RWMutex
	topicUnpackedTxns map[string]IunpackedTxns

	baseChecks   []TxnCheckFn
	tripodChecks map[string]TxnCheckFn

	filter func(txn *SignedTxn) bool
}

func NewTxPool(nodeType int, cfg *TxpoolConf) *TxPool {
	ordered := newOrderedTxns()

	tp := &TxPool{
		nodeType:          nodeType,
		capacity:          cfg.PoolSize,
		TxnMaxSize:        cfg.TxnMaxSize,
		unpackedTxns:      ordered,
		topicUnpackedTxns: make(map[string]IunpackedTxns),
		baseChecks:        make([]TxnCheckFn, 0),
		tripodChecks:      make(map[string]TxnCheckFn),
		filter:            func(*SignedTxn) bool { return true },
	}
	return tp
}

func WithDefaultChecks(nodeType int, cfg *TxpoolConf) *TxPool {
	tp := NewTxPool(nodeType, cfg)
	return tp.withDefaultBaseChecks()
}

func (tp *TxPool) withDefaultBaseChecks() *TxPool {
	tp.baseChecks = []TxnCheckFn{
		tp.checkPoolLimit,
		tp.checkTxnSize,
	}
	return tp
}

func (tp *TxPool) SetPackFilter(fn func(txn *SignedTxn) bool) {
	tp.filter = fn
}

func (tp *TxPool) Capacity() int {
	return tp.capacity
}

func (tp *TxPool) Size() int {
	return tp.unpackedTxns.Size()
}

func (tp *TxPool) WithBaseCheck(tc TxnChecker) ItxPool {
	tp.baseChecks = append(tp.baseChecks, tc.CheckTxn)
	return tp
}

func (tp *TxPool) WithTripodCheck(tripodName string, tc TxnChecker) ItxPool {
	tp.tripodChecks[tripodName] = tc.CheckTxn
	return tp
}

func (tp *TxPool) Exist(txnHash Hash) bool {
	//tp.RLock()
	//defer tp.RUnlock()
	return tp.unpackedTxns.Exist(txnHash)
}

func (tp *TxPool) CheckTxn(stxn *SignedTxn) (err error) {
	err = tp.BaseCheck(stxn)
	if err != nil {
		return
	}
	return tp.TripodsCheck(stxn)
}

func (tp *TxPool) Insert(stxn *SignedTxn) error {
	metrics.TxPoolInsertCounter.WithLabelValues().Inc()
	if tp.nodeType == LightNode {
		return nil
	}
	return tp.unpackedTxns.Insert(stxn)
}

func (tp *TxPool) InsertWithTopic(topic string, stxn *SignedTxn) error {
	tp.topicLock.Lock()
	defer tp.topicLock.Unlock()
	if _, ok := tp.topicUnpackedTxns[topic]; !ok {
		tp.topicUnpackedTxns[topic] = newOrderedTxns()
	}
	return tp.topicUnpackedTxns[topic].Insert(stxn)
}

func (tp *TxPool) PackWithTopic(topic string, numLimit uint64) ([]*SignedTxn, error) {
	if unpack, ok := tp.topicUnpackedTxns[topic]; ok {
		return unpack.Gets(numLimit, tp.filter), nil
	}
	return nil, nil
}

func (tp *TxPool) PackWithTopicFor(topic string, numLimit uint64, filter func(txn *SignedTxn) bool) ([]*SignedTxn, error) {
	tp.topicLock.RLock()
	defer tp.topicLock.RUnlock()
	if unpack, ok := tp.topicUnpackedTxns[topic]; ok {
		return unpack.Gets(numLimit, filter), nil
	}
	return nil, nil
}

func (tp *TxPool) SetOrder(order map[int]Hash) {
	//tp.Lock()
	//defer tp.Unlock()
	tp.unpackedTxns.SetOrder(order)
}

func (tp *TxPool) SortTxns(fn func(txns []*SignedTxn) []*SignedTxn) {
	//tp.Lock()
	//defer tp.Unlock()
	tp.unpackedTxns.SortTxns(fn)
}

func (tp *TxPool) GetTxn(hash Hash) (*SignedTxn, error) {
	//tp.RLock()
	//defer tp.RUnlock()
	return tp.unpackedTxns.Get(hash), nil
}

func (tp *TxPool) GetAllTxns() ([]*SignedTxn, error) {
	//tp.RLock()
	//defer tp.RUnlock()
	return tp.unpackedTxns.GetAll(), nil
}

func (tp *TxPool) Pack(numLimit uint64) ([]*SignedTxn, error) {
	metrics.TxpoolSizeGauge.Set(float64(tp.unpackedTxns.Size()))
	txns := tp.unpackedTxns.Gets(numLimit, tp.filter)
	return txns, nil
}

func (tp *TxPool) PackFor(numLimit uint64, filter func(txn *SignedTxn) bool) ([]*SignedTxn, error) {
	//tp.RLock()
	//defer tp.RUnlock()
	metrics.TxpoolSizeGauge.Set(float64(tp.unpackedTxns.Size()))
	txns := tp.unpackedTxns.Gets(numLimit, filter)
	return txns, nil
}

func (tp *TxPool) Reset(txns SignedTxns) error {
	//tp.Lock()
	//defer tp.Unlock()
	tp.unpackedTxns.Deletes(txns.Hashes())
	return nil
}

func (tp *TxPool) ResetByHashes(hashes []Hash) error {
	//tp.Lock()
	//defer tp.Unlock()
	tp.unpackedTxns.Deletes(hashes)
	return nil
}

func (tp *TxPool) ResetByHashesAndTopic(hashes []Hash, topic string) error {
	tp.topicLock.Lock()
	defer tp.topicLock.Unlock()
	if unpack, ok := tp.topicUnpackedTxns[topic]; ok {
		unpack.Deletes(hashes)
	}
	return nil
}

// ------------------- check txn rules ----------------------

func (tp *TxPool) BaseCheck(stxn *SignedTxn) error {
	return Check(tp.baseChecks, stxn)
}

func (tp *TxPool) TripodsCheck(stxn *SignedTxn) error {
	tripodCheck := tp.tripodChecks[stxn.TripodName()]
	return tripodCheck(stxn)
}

func (tp *TxPool) NecessaryCheck(stxn *SignedTxn) (err error) {
	err = tp.checkTxnSize(stxn)
	if err != nil {
		return
	}

	return tp.TripodsCheck(stxn)
}

func (tp *TxPool) checkPoolLimit(*SignedTxn) error {
	if tp.unpackedTxns.Size() >= tp.capacity {
		return PoolOverflow
	}
	return nil
}

func (tp *TxPool) checkTxnSize(stxn *SignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return nil
}

type TxnCheckFn func(*SignedTxn) error

func Check(checks []TxnCheckFn, stxn *SignedTxn) error {
	for _, check := range checks {
		err := check(stxn)
		if err != nil {
			return err
		}
	}
	return nil
}

//func CheckSignature(stxn *SignedTxn) error {
//	sig := stxn.Signature
//	wrCall := stxn.Raw.WrCall
//	hash, err := wrCall.Hash()
//	if err != nil {
//		return err
//	}
//	if !stxn.Pubkey.VerifySignature(hash, sig) {
//		return TxnSignatureErr
//	}
//	return nil
//}
