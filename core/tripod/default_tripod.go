package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

type DefaultTripod struct {
	*TripodHeader
}

func NewDefaultTripod(name string) *DefaultTripod {
	meta := NewTripodHeader(name)
	dt := &DefaultTripod{
		TripodHeader: meta,
	}
	dt.SetBlockCycle(dt)
	dt.SetTxnChecker(dt)
	dt.SetBlockVerifier(dt)
	return dt
}

func (dt *DefaultTripod) GetTripodHeader() *TripodHeader {
	return dt.TripodHeader
}

func (dt *DefaultTripod) CheckTxn(*SignedTxn) error {
	return nil
}

func (dt *DefaultTripod) VerifyBlock(*Block) bool {
	return true
}

func (*DefaultTripod) InitChain() {
}

func (*DefaultTripod) SyncHistory() {
}

func (*DefaultTripod) StartBlock(*Block) {
}

func (*DefaultTripod) EndBlock(*Block) {
}

func (*DefaultTripod) FinalizeBlock(*Block) {
}
