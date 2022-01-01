package txpool

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

type orderedTxns struct {
	index map[Hash]int
	txns  []*SignedTxn
}

func newOrderedTxns() *orderedTxns {
	return &orderedTxns{
		index: make(map[Hash]int),
		txns:  make([]*SignedTxn, 0),
	}
}

func (ot *orderedTxns) insertTx(input *SignedTxn) {
	if _, ok := ot.index[input.TxnHash]; ok {
		return
	}
	if len(ot.txns) == 0 {
		ot.txns = []*SignedTxn{input}
	}
	for i, tx := range ot.txns {
		if input.Raw.LeiPrice > tx.Raw.LeiPrice {
			ot.txns = append(ot.txns[:i], append([]*SignedTxn{input}, ot.txns[i:]...)...)
			ot.index[input.TxnHash] = i
			return
		}
	}
}

func (ot *orderedTxns) delete(hash Hash) {
	if idx, ok := ot.index[hash]; ok {
		return
	} else {
		ot.txns = append(ot.txns[:idx], ot.txns[idx+1:]...)
		delete(ot.index, hash)
	}
}

func (ot *orderedTxns) pop(count int) []*SignedTxn {
	pops := ot.txns[:count]
	ot.txns = ot.txns[count:]
	for _, pop := range pops {
		delete(ot.index, pop.TxnHash)
	}
	return pops
}
