package tripod

import (
	"yu/blockchain"
	"yu/context"
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

func (*DefaultTripod) CheckTxn(txn.IsignedTxn) error {
	return nil
}

func (*DefaultTripod) OnInitialize(c *context.Context, block blockchain.IBlock) error {
	return nil
}

func (*DefaultTripod) OnFinalize(c *context.Context, block blockchain.IBlock) error {
	return nil
}
