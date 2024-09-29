package txpool

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/metrics"
)

type TxPool struct {
	// sync.RWMutex

	nodeType int

	capacity   int
	TxnMaxSize int

	unpackedTxns IunpackedTxns

	baseChecks   []TxnCheckFn
	tripodChecks map[string]TxnCheckFn
}

func NewTxPool(nodeType int, cfg *TxpoolConf) *TxPool {
	ordered := newOrderedTxns()

	tp := &TxPool{
		nodeType:     nodeType,
		capacity:     cfg.PoolSize,
		TxnMaxSize:   cfg.TxnMaxSize,
		unpackedTxns: ordered,
		baseChecks:   make([]TxnCheckFn, 0),
		tripodChecks: make(map[string]TxnCheckFn),
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
	tp.unpackedTxns.Insert(stxn)
	return nil
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
	return tp.PackFor(numLimit, func(*SignedTxn) bool {
		return true
	})
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
