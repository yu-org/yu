package txpool

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
	"sort"
)

type orderedTxns struct {
	txns []*SignedTxn
	idx  map[Hash]*SignedTxn

	// order map[Hash]int
}

func newOrderedTxns() *orderedTxns {
	return &orderedTxns{
		txns: make([]*SignedTxn, 0),
		idx:  make(map[Hash]*SignedTxn),
		// order: make(map[Hash]int),
	}
}

func (ot *orderedTxns) Insert(input *SignedTxn) {
	logrus.WithField("txpool", "ordered-txns").
		Tracef("Insert txn(%s) to Txpool, txn content: %v", input.TxnHash, input.Raw.WrCall)

	ot.idx[input.TxnHash] = input
	ot.txns = append(ot.txns, input)

	//for element := ot.txns.Front(); element != nil; element = element.Next() {
	//	tx := element.Value.(*SignedTxn)
	//	// fixme: cannot only use tips to judge.
	//	if input.Raw.WrCall.Tips > tx.Raw.WrCall.Tips {
	//		e := ot.txns.InsertBefore(input, element)
	//		ot.idx[input.TxnHash] = e
	//		return
	//	}
	//}
	//e := ot.txns.PushBack(input)
	//ot.idx[input.TxnHash] = e
}

func (ot *orderedTxns) SetOrder(order map[Hash]int) {
	panic("implement me")
	// ot.order = order
}

func (ot *orderedTxns) SortBy(lessFunc func(i, j int) bool) {
	sort.Slice(ot.txns, lessFunc)
}

func (ot *orderedTxns) delete(txnHash Hash) {
	if stxn, ok := ot.idx[txnHash]; ok {
		logrus.WithField("txpool", "ordered-txns").
			Tracef("DELETE txn(%s) from txpool, txn content: %v", stxn.TxnHash.String(), stxn.Raw.WrCall)
		delete(ot.idx, txnHash)
		// delete(ot.order, txnHash)
	}
	for i, txn := range ot.txns {
		if txn.TxnHash == txnHash {
			ot.txns = append(ot.txns[:i], ot.txns[i+1:]...)
		}
	}
}

func (ot *orderedTxns) Deletes(txnHashes []Hash) {
	for _, hash := range txnHashes {
		ot.delete(hash)
	}
}

func (ot *orderedTxns) Exist(txnHash Hash) bool {
	_, ok := ot.idx[txnHash]
	return ok
}

func (ot *orderedTxns) Get(txnHash Hash) *SignedTxn {
	return ot.idx[txnHash]
}

func (ot *orderedTxns) Gets(numLimit uint64, filter func(txn *SignedTxn) bool) []*SignedTxn {
	if numLimit > uint64(ot.Size()) {
		numLimit = uint64(ot.Size())
	}

	txns := make([]*SignedTxn, numLimit)

	//for hash, num := range ot.order {
	//	txns[num]
	//}

	for i := 0; i < int(numLimit); i++ {

		txn := ot.txns[i]
		if filter(txn) {
			logrus.WithField("txpool", "ordered-txns").
				Tracef("Pack txn(%s) from Txpool, txn content: %v", txn.TxnHash, txn.Raw.WrCall)
			txns[i] = txn
		}
	}
	return txns
}

func (ot *orderedTxns) GetAll() []*SignedTxn {
	return ot.txns
}

func (ot *orderedTxns) Size() int {
	return len(ot.txns)
}
