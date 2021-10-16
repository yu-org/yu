package tripod

import (
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	"github.com/yu-org/yu/txn"
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

func (*DefaultTripod) CheckTxn(*txn.SignedTxn) error {
	return nil
}

func (*DefaultTripod) VerifyBlock(IBlock) bool {
	return true
}

func (*DefaultTripod) InitChain() error {
	return nil
}

func (*DefaultTripod) StartBlock(IBlock) error {
	return nil
}

func (*DefaultTripod) EndBlock(IBlock) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(IBlock) error {
	return nil
}
