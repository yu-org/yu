package txpool

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

type orderedTxns struct {
	txns []*SignedTxn
}

func newOrderedTxns() *orderedTxns {
	return &orderedTxns{
		txns: make([]*SignedTxn, 0),
	}
}

func (ot *orderedTxns) insert(input *SignedTxn) {
	if len(ot.txns) == 0 {
		ot.txns = []*SignedTxn{input}
	}
	for i, tx := range ot.txns {
		if tx == nil {
			continue
		}
		if input.Raw.Ecall.LeiPrice > tx.Raw.Ecall.LeiPrice {
			ot.txns = append(ot.txns[:i], append([]*SignedTxn{input}, ot.txns[i:]...)...)
			return
		}
	}
}

func (ot *orderedTxns) delete(hash Hash) {
	for idx, txn := range ot.txns {
		if txn.TxnHash == hash {
			logrus.Tracef("DELETE txn(%s) from txpool", hash.String())
			ot.txns = append(ot.txns[:idx], ot.txns[idx+1:]...)
			return
		}
	}
}

func (ot *orderedTxns) deletes(hashes []Hash) {
	for _, hash := range hashes {
		ot.delete(hash)
	}
}

func (ot *orderedTxns) gets(numLimit uint64, filter func(txn *SignedTxn) bool) []*SignedTxn {
	txns := make([]*SignedTxn, 0)
	if numLimit > uint64(ot.size()) {
		numLimit = uint64(ot.size())
	}
	for _, txn := range ot.txns[:numLimit] {
		if filter(txn) && txn != nil {
			logrus.Tracef("Pack txn(%s) from Txpool", txn.TxnHash.String())
			txns = append(txns, txn)
		}
	}
	return txns
}

func (ot *orderedTxns) size() int {
	return len(ot.txns)
}
