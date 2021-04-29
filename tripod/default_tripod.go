package tripod

import (
	. "yu/blockchain"
	"yu/txn"
	. "yu/txpool"
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

func (*DefaultTripod) ValidateBlock(IBlockChain, IBlock) bool {
	return false
}

func (*DefaultTripod) InitChain(IBlockChain, IBlockBase) error {
	return nil
}

func (*DefaultTripod) StartBlock(IBlockChain, IBlock, ItxPool) (bool, error) {
	return false, nil
}

func (*DefaultTripod) EndBlock(IBlockChain, IBlock, ItxPool) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(IBlockChain, IBlock) error {
	return nil
}
