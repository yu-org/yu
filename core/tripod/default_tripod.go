package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

type DefaultTripod struct {
	*Tripod
}

func NewDefaultTripod(name string) *DefaultTripod {
	meta := NewTripod(name)
	dt := &DefaultTripod{
		Tripod: meta,
	}
	dt.SetInit(dt)
	dt.SetBlockCycle(dt)
	dt.SetTxnChecker(dt)
	dt.SetBlockVerifier(dt)
	return dt
}

func (dt *DefaultTripod) GetTripodHeader() *Tripod {
	return dt.Tripod
}

func (dt *DefaultTripod) CheckTxn(*SignedTxn) error {
	return nil
}

func (dt *DefaultTripod) VerifyBlock(*Block) bool {
	return true
}

func (*DefaultTripod) InitChain() {
}

func (*DefaultTripod) StartBlock(*Block) {
}

func (*DefaultTripod) EndBlock(*Block) {
}

func (*DefaultTripod) FinalizeBlock(*Block) {
}
