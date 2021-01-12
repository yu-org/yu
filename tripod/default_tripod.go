package tripod

import "yu/blockchain"

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

func (dt *DefaultTripod) OnInitialize(block blockchain.IBlock) error {
	return nil
}

func (dt *DefaultTripod) OnFinalize(block blockchain.IBlock) error {
	return nil
}
