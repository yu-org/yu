package poa

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/chain_env"
	. "github.com/Lawliet-Chan/yu/tripod"
	. "github.com/Lawliet-Chan/yu/txn"
)

type Poa struct {
	meta *TripodMeta
}

func NewPoa() *Poa {
	meta := NewTripodMeta("poa")

	return &Poa{meta: meta}
}

func (p *Poa) GetTripodMeta() *TripodMeta {
	return p.meta
}

func (p *Poa) Name() string {
	return p.meta.Name()
}

func (p *Poa) CheckTxn(txn *SignedTxn) error {
	return nil
}

func (p *Poa) VerifyBlock(block IBlock, env *ChainEnv) bool {
	panic("implement me")
}

func (p *Poa) InitChain(env *ChainEnv, _ *Land) error {
	chain := env.Chain
	gensisBlock := &Block{
		Header: &Header{},
	}
	return chain.SetGenesis(gensisBlock)
}

func (p *Poa) StartBlock(block IBlock, env *ChainEnv, land *Land) (needBroadcast bool, err error) {
	panic("implement me")
}

func (p *Poa) EndBlock(block IBlock, env *ChainEnv, land *Land) error {
	panic("implement me")
}

func (p *Poa) FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error {
	panic("implement me")
}
