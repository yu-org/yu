package tripod

import (
	. "yu/blockchain"
	. "yu/env"
	"yu/txn"
)

type DefaultTripod struct {
	meta *TripodMeta
}

func NewDefaultTripod(name string) *DefaultTripod {
	meta := NewTripodMeta(name)
	return &DefaultTripod{
		meta: meta,
	}
}

func (dt *DefaultTripod) TripodMeta() *TripodMeta {
	return dt.meta
}

func (*DefaultTripod) CheckTxn(*txn.SignedTxn) error {
	return nil
}

func (*DefaultTripod) ValidateBlock(IBlock, *Env) bool {
	return false
}

func (*DefaultTripod) InitChain(*Env, *Land) error {
	return nil
}

func (*DefaultTripod) StartBlock(*Env, *Land) (IBlock, bool, error) {
	return nil, false, nil
}

func (*DefaultTripod) EndBlock(IBlock, *Env, *Land) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(IBlock, *Env, *Land) error {
	return nil
}
