package txpool

import (
	"container/list"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

type orderedTxns struct {
	txns *list.List
	idx  map[Hash]*list.Element
}

func newOrderedTxns() *orderedTxns {
	return &orderedTxns{
		txns: list.New(),
		idx:  make(map[Hash]*list.Element),
	}
}

func (ot *orderedTxns) Insert(input *SignedTxn) {
	logrus.WithField("txpool", "ordered-txns").
		Tracef("Insert txn(%s) to Txpool, txn content: %v", input.TxnHash.String(), input.Raw.WrCall)
	for element := ot.txns.Front(); element != nil; element = element.Next() {
		tx := element.Value.(*SignedTxn)
		if input.Raw.WrCall.LeiPrice > tx.Raw.WrCall.LeiPrice {
			e := ot.txns.InsertBefore(input, element)
			ot.idx[input.TxnHash] = e
			return
		}
	}
	e := ot.txns.PushBack(input)
	ot.idx[input.TxnHash] = e
}

func (ot *orderedTxns) delete(txnHash Hash) {
	if e, ok := ot.idx[txnHash]; ok {
		stxn := e.Value.(*SignedTxn)
		logrus.WithField("txpool", "ordered-txns").
			Tracef("DELETE txn(%s) from txpool, txn content: %v", stxn.TxnHash.String(), stxn.Raw.WrCall)
		ot.txns.Remove(e)
		delete(ot.idx, txnHash)
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
	if e, ok := ot.idx[txnHash]; ok {
		return e.Value.(*SignedTxn)
	}
	return nil
}

func (ot *orderedTxns) Gets(numLimit uint64, filter func(txn *SignedTxn) bool) []*SignedTxn {
	txns := make([]*SignedTxn, 0)
	if numLimit > uint64(ot.Size()) {
		numLimit = uint64(ot.Size())
	}
	var packedNum uint64 = 0
	for element := ot.txns.Front(); element != nil && packedNum < numLimit; element = element.Next() {
		txn := element.Value.(*SignedTxn)
		if filter(txn) {
			logrus.WithField("txpool", "ordered-txns").
				Tracef("Pack txn(%s) from Txpool, txn content: %v", txn.TxnHash.String(), txn.Raw.WrCall)
			txns = append(txns, txn)
			packedNum++
		}
	}
	return txns
}

func (ot *orderedTxns) Size() int {
	return ot.txns.Len()
}
