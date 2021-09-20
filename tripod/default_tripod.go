package tripod

import (
	. "github.com/yu-altar/yu/blockchain"
	. "github.com/yu-altar/yu/chain_env"
	"github.com/yu-altar/yu/txn"
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

func (*DefaultTripod) StartBlock(IBlock, *ChainEnv, *Land, <-chan []byte) ([]byte, error) {
	return nil, nil
}

func (*DefaultTripod) EndBlock(IBlock, *ChainEnv, *Land, <-chan []byte) ([]byte, error) {
	return nil, nil
}

func (*DefaultTripod) FinalizeBlock(IBlock, *ChainEnv, *Land, <-chan []byte) ([]byte, error) {
	return nil, nil
}
