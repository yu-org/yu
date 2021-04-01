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

func (*DefaultTripod) CheckTxn(*txn.SignedTxn) error {
	return nil
}

func (*DefaultTripod) ValidateBlock(IBlock) bool {
	return false
}

func (*DefaultTripod) StartBlock(IBlockChain, IBlock, txpool.ItxPool) (bool, error) {
	return false, nil
}

func (*DefaultTripod) EndBlock(IBlockChain, IBlock) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(IBlockChain, IBlock) error {
	return nil
}
