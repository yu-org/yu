package tripod

import (
	. "github.com/yu-org/yu/chain_env"
	"github.com/yu-org/yu/types"
)

type DefaultTripod struct {
	*TripodMeta
	*ChainEnv
}

func NewDefaultTripod(name string) *DefaultTripod {
	meta := NewTripodMeta(name)
	return &DefaultTripod{
		TripodMeta: meta,
	}
}

func (dt *DefaultTripod) GetTripodMeta() *TripodMeta {
	return dt.TripodMeta
}

func (dt *DefaultTripod) SetChainEnv(env *ChainEnv) {
	dt.ChainEnv = env
}

func (*DefaultTripod) CheckTxn(*types.SignedTxn) error {
	return nil
}

func (*DefaultTripod) VerifyBlock(types.IBlock) bool {
	return true
}

func (*DefaultTripod) InitChain() error {
	return nil
}

func (*DefaultTripod) StartBlock(types.IBlock) error {
	return nil
}

func (*DefaultTripod) EndBlock(types.IBlock) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(types.IBlock) error {
	return nil
}
