package txpool

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

type orderedTxns struct {
	sync.RWMutex
	txns []*SignedTxn
	idx  map[Hash]*SignedTxn

	// order map[int]Hash
}

func newOrderedTxns() *orderedTxns {
	return &orderedTxns{
		txns: make([]*SignedTxn, 0),
		idx:  make(map[Hash]*SignedTxn),
		// order: make(map[int]Hash),
	}
}

func (ot *orderedTxns) Insert(input *SignedTxn) error {
	//logrus.WithField("txpool", "ordered-txns").
	//	Tracef("Insert txn(%s) to Txpool, txn content: %v", input.TxnHash, input.Raw.WrCall)
	TxnPoolCounter.WithLabelValues("insert").Inc()
	start := time.Now()
	defer func() {
		TxnPoolDuration.WithLabelValues("insert").Observe(float64(time.Since(start).Microseconds()))
	}()
	ot.Lock()
	defer ot.Unlock()
	if _, ok := ot.idx[input.TxnHash]; ok {
		return fmt.Errorf("insert txn %s duplicated", input.TxnHash.String())
	}
	ot.idx[input.TxnHash] = input
	ot.txns = append(ot.txns, input)
	return nil
}

func (ot *orderedTxns) SetOrder(order map[int]Hash) {
	panic("implement me")
	//for i, hash := range order {
	//	ot.setOrder(i, hash)
	//}
}

//func (ot *orderedTxns) setOrder(num int, txHash Hash) {
//	if _, ok := ot.order[num]; ok {
//		ot.setOrder(num+1, txHash)
//	}
//	ot.order[num] = txHash
//}

func (ot *orderedTxns) Deletes(txnHashes []Hash) {
	TxnPoolCounter.WithLabelValues("deletes").Inc()
	start := time.Now()
	defer func() {
		TxnPoolDuration.WithLabelValues("deletes").Observe(float64(time.Since(start).Microseconds()))
	}()
	txnMap := make(map[Hash]struct{})
	for _, txnHash := range txnHashes {
		txnMap[txnHash] = struct{}{}
	}
	ot.Lock()
	defer ot.Unlock()
	for txnHash := range txnMap {
		delete(ot.idx, txnHash)
	}
	for i := 0; i < len(ot.txns) && len(txnMap) > 0; i++ {
		_, ok := txnMap[ot.txns[i].TxnHash]
		if ok {
			delete(txnMap, ot.txns[i].TxnHash)
			ot.txns = append(ot.txns[:i], ot.txns[i+1:]...)
			i--
		}
	}
}

func (ot *orderedTxns) Exist(txnHash Hash) bool {
	TxnPoolCounter.WithLabelValues("exist").Inc()
	start := time.Now()
	defer func() {
		TxnPoolDuration.WithLabelValues("exist").Observe(float64(time.Since(start).Microseconds()))
	}()
	ot.RLock()
	defer ot.RUnlock()
	_, ok := ot.idx[txnHash]
	return ok
}

func (ot *orderedTxns) Get(txnHash Hash) *SignedTxn {
	TxnPoolCounter.WithLabelValues("get").Inc()
	start := time.Now()
	defer func() {
		TxnPoolDuration.WithLabelValues("get").Observe(float64(time.Since(start).Microseconds()))
	}()
	ot.RLock()
	defer ot.RUnlock()
	return ot.idx[txnHash]
}

func (ot *orderedTxns) Gets(numLimit uint64, filter func(txn *SignedTxn) bool) []*SignedTxn {
	TxnPoolCounter.WithLabelValues("gets").Inc()
	start := time.Now()
	defer func() {
		TxnPoolDuration.WithLabelValues("gets").Observe(float64(time.Since(start).Microseconds()))
	}()
	ot.RLock()
	defer ot.RUnlock()
	size := ot.size()
	if numLimit > uint64(size) {
		numLimit = uint64(size)
	}

	txns := make([]*SignedTxn, 0)

	//ot.excludeEmptyOrder()
	//
	//for num, hash := range ot.order {
	//	// FIXME: if num > numLimit, panic here
	//	if filter(ot.idx[hash]) {
	//		txns[num] = ot.idx[hash]
	//	}
	//}

	for i := 0; i < int(numLimit); i++ {
		//if txns[i] != nil {
		//	continue
		//}

		txn := ot.txns[i]

		if filter(txn) {
			logrus.WithField("txpool", "ordered-txns").
				Tracef("Pack txn(%s) from Txpool, txn content: %v", txn.TxnHash, txn.Raw.WrCall)
			txns = append(txns, txn)
			// txns[i] = txn
		}
	}
	return txns
}

//func (ot *orderedTxns) excludeEmptyOrder() {
//	for num, hash := range ot.order {
//		if _, ok := ot.idx[hash]; !ok {
//			delete(ot.order, num)
//		}
//	}
//}

func (ot *orderedTxns) SortTxns(fn func(txns []*SignedTxn) []*SignedTxn) {
	TxnPoolCounter.WithLabelValues("sorttxns").Inc()
	start := time.Now()
	defer func() {
		TxnPoolDuration.WithLabelValues("sorttxns").Observe(float64(time.Since(start).Microseconds()))
	}()
	ot.Lock()
	orderedTxs := fn(ot.txns)
	ot.txns = orderedTxs
	ot.Unlock()
}

func (ot *orderedTxns) GetAll() []*SignedTxn {
	TxnPoolCounter.WithLabelValues("getAll").Inc()
	start := time.Now()
	defer func() {
		TxnPoolDuration.WithLabelValues("getAll").Observe(float64(time.Since(start).Microseconds()))
	}()
	ot.RLock()
	defer ot.RUnlock()
	txns := make([]*SignedTxn, 0, len(ot.txns))
	for _, txn := range ot.txns {
		txns = append(txns, txn)
	}
	return txns
}

func (ot *orderedTxns) Size() int {
	TxnPoolCounter.WithLabelValues("size").Inc()
	start := time.Now()
	defer func() {
		TxnPoolDuration.WithLabelValues("size").Observe(float64(time.Since(start).Microseconds()))
	}()
	ot.RLock()
	defer ot.RUnlock()
	return ot.size()
}

func (ot *orderedTxns) size() int {
	return len(ot.txns)
}
