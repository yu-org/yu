package poa

import (
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/tripod"
	. "github.com/yu-org/yu/txn"
)

type Poa struct {
}

func NewPoa() *Poa {

}

func (p *Poa) GetTripodMeta() *TripodMeta {
	panic("implement me")
}

func (p *Poa) Name() string {
	return "poa"
}

func (p *Poa) CheckTxn(txn *SignedTxn) error {
	panic("implement me")
}

func (p *Poa) VerifyBlock(block IBlock, env *ChainEnv) bool {
	panic("implement me")
}

func (p *Poa) InitChain(env *ChainEnv, land *Land) error {
	panic("implement me")
}

func (p *Poa) StartBlock(block IBlock, env *ChainEnv, land *Land) error {
	panic("implement me")
}

func (p *Poa) EndBlock(block IBlock, env *ChainEnv, land *Land) error {
	panic("implement me")
}

func (p *Poa) FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error {
	panic("implement me")
}
