package tripod

import (
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

func (*DefaultTripod) ValidateBlock(*Env) bool {
	return false
}

func (*DefaultTripod) InitChain(*Env, *Land) error {
	return nil
}

func (*DefaultTripod) StartBlock(*Env, *Land) (bool, error) {
	return false, nil
}

func (*DefaultTripod) EndBlock(*Env, *Land) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(*Env, *Land) error {
	return nil
}
