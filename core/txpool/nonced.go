package txpool

import (
	"container/list"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

type noncedTxns struct {
	txns map[Address]*list.List
	idx  map[Hash]*list.Element
}

func newNoncedTxns() *noncedTxns {
	return &noncedTxns{
		txns: make(map[Address]*list.List),
		idx:  make(map[Hash]*list.Element),
	}
}

func (n *noncedTxns) Insert(input *SignedTxn) {
	//TODO implement me
	panic("implement me")
}

func (n *noncedTxns) Deletes(txnHashes []Hash) {
	//TODO implement me
	panic("implement me")
}

func (n *noncedTxns) Exist(txnHash Hash) bool {
	//TODO implement me
	panic("implement me")
}

func (n *noncedTxns) Get(txnHash Hash) *SignedTxn {
	//TODO implement me
	panic("implement me")
}

func (n *noncedTxns) Gets(numLimit uint64, filter func(txn *SignedTxn) bool) []*SignedTxn {
	//TODO implement me
	panic("implement me")
}

func (n *noncedTxns) Size() int {
	//TODO implement me
	panic("implement me")
}
