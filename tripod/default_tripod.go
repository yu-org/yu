package tripod

import (
	. "yu/blockchain"
	"yu/txn"
	"yu/txpool"
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

func (*DefaultTripod) StartBlock(IBlockChain, IBlock, txpool.ItxPool) error {
	return nil
}

func (*DefaultTripod) EndBlock(IBlock) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(IBlockChain, IBlock) error {
	return nil
}
