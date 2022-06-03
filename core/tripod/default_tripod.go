package tripod

import (
	. "github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
)

type DefaultTripod struct {
	*TripodHeader

	txnChecker    TxnCheckFn
	blockVerifier BlockVerifier
}

type BlockVerifier func(block *CompactBlock) bool

func NewDefaultTripod(name string) *DefaultTripod {
	meta := NewTripodHeader(name)
	return &DefaultTripod{
		TripodHeader: meta,
	}
}

func (dt *DefaultTripod) SetTxnChecker(checker TxnCheckFn) {
	dt.txnChecker = checker
}

func (dt *DefaultTripod) SetBlockVerifier(fn BlockVerifier) {
	dt.blockVerifier = fn
}

func (dt *DefaultTripod) GetTripodHeader() *TripodHeader {
	return dt.TripodHeader
}

func (dt *DefaultTripod) CheckTxn(txn *SignedTxn) error {
	if dt.txnChecker == nil {
		return nil
	}
	return dt.txnChecker(txn)
}

func (dt *DefaultTripod) VerifyBlock(block *CompactBlock) bool {
	if dt.blockVerifier == nil {
		return true
	}
	return dt.blockVerifier(block)
}

func (*DefaultTripod) InitChain() {
}

func (*DefaultTripod) StartBlock(*CompactBlock) {
}

func (*DefaultTripod) EndBlock(*CompactBlock) {
}

func (*DefaultTripod) FinalizeBlock(*CompactBlock) {
}
