package tripod

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/chain_env"
	"github.com/Lawliet-Chan/yu/txn"
)

type DefaultTripod struct {
	*DefaultChainLifeCycle
	Meta *TripodMeta
}

func NewDefaultTripod(name string) *DefaultTripod {
	meta := NewTripodMeta(name)
	return &DefaultTripod{
		DefaultChainLifeCycle: &DefaultChainLifeCycle{},
		Meta:                  meta,
	}
}

func (dt *DefaultTripod) TripodMeta() *TripodMeta {
	return dt.Meta
}

func (dt *DefaultTripod) Name() string {
	return dt.Meta.name
}

func (*DefaultTripod) CheckTxn(*txn.SignedTxn) error {
	return nil
}

func (*DefaultTripod) ValidateBlock(IBlock, *ChainEnv) bool {
	return false
}

type DefaultChainLifeCycle struct{}

func (*DefaultChainLifeCycle) InitChain(*ChainEnv, *Land) error {
	return nil
}

func (*DefaultChainLifeCycle) StartBlock(*ChainEnv, *Land) (IBlock, bool, error) {
	return nil, false, nil
}

func (*DefaultChainLifeCycle) EndBlock(IBlock, *ChainEnv, *Land) error {
	return nil
}

func (*DefaultChainLifeCycle) FinalizeBlock(IBlock, *ChainEnv, *Land) error {
	return nil
}
