package poa

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/chain_env"
	. "github.com/Lawliet-Chan/yu/consensus/hotstuff"
	. "github.com/Lawliet-Chan/yu/tripod"
	. "github.com/Lawliet-Chan/yu/txn"
)

type Poa struct {
	meta *TripodMeta

	smr *Smr
}

func NewPoa(addr string, addrs []string) *Poa {
	meta := NewTripodMeta("poa")

	q := InitQcTee()
	saftyrules := &DefaultSaftyRules{
		QcTree: q,
	}
	elec := NewSimpleElection(addrs)
	smr := NewSmr(addr, &DefaultPaceMaker{}, saftyrules, elec, q)

	return &Poa{
		meta: meta,
		smr:  smr,
	}
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

func InitQcTee() *QCPendingTree {
	initQC := &QuorumCert{
		VoteInfo: &VoteInfo{
			ProposalId:   []byte{0},
			ProposalView: 0,
		},
		LedgerCommitInfo: &LedgerCommitInfo{
			CommitStateId: []byte{0},
		},
	}
	rootNode := &ProposalNode{
		In: initQC,
	}
	return &QCPendingTree{
		Genesis:  rootNode,
		Root:     rootNode,
		HighQC:   rootNode,
		CommitQC: rootNode,
	}
}
