package tripod

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/chain_env"
	"github.com/Lawliet-Chan/yu/txn"
)

type DefaultTripod struct {
	*TripodMeta
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

func (*DefaultTripod) CheckTxn(*txn.SignedTxn) error {
	return nil
}

func (*DefaultTripod) VerifyBlock(IBlock, *ChainEnv) bool {
	return true
}

func (*DefaultTripod) InitChain(*ChainEnv, *Land) error {
	return nil
}

func (*DefaultTripod) StartBlock(IBlock, *ChainEnv, *Land) (bool, error) {
	return false, nil
}

func (*DefaultTripod) EndBlock(IBlock, *ChainEnv, *Land) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(IBlock, *ChainEnv, *Land) error {
	return nil
}
