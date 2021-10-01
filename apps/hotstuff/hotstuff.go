package hotstuff

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/consensus/chained-hotstuff"
	. "github.com/yu-org/yu/tripod"
	. "github.com/yu-org/yu/txn"
)

type Hotstuff struct {
	meta         *TripodMeta
	validatorsIP []string

	smr *Smr
}

func NewHotstuff(addr string, validatorsIP []string) *Hotstuff {
	meta := NewTripodMeta("hotstuff")

	q := InitQcTee()
	saftyrules := &DefaultSaftyRules{
		QcTree: q,
	}
	elec := NewSimpleElection(validatorsIP)
	smr := NewSmr(addr, &DefaultPaceMaker{}, saftyrules, elec, q)

	return &Hotstuff{
		meta:         meta,
		validatorsIP: validatorsIP,
		smr:          smr,
	}
}

func (p *Hotstuff) GetTripodMeta() *TripodMeta {
	return p.meta
}

func (p *Hotstuff) Name() string {
	return p.meta.Name()
}

func (p *Hotstuff) CheckTxn(txn *SignedTxn) error {
	return nil
}

func (p *Hotstuff) VerifyBlock(block IBlock, env *ChainEnv) bool {
	return true
}

func (p *Hotstuff) InitChain(env *ChainEnv, _ *Land) error {
	chain := env.Chain
	gensisBlock := &Block{
		Header: &Header{},
	}
	return chain.SetGenesis(gensisBlock)
}

func (p *Hotstuff) StartBlock(block IBlock, env *ChainEnv, land *Land) (needBroadcast bool, err error) {
	panic("implement me")
}

func (p *Hotstuff) EndBlock(block IBlock, env *ChainEnv, land *Land) error {
	panic("implement me")
}

func (p *Hotstuff) FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error {
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

func (p *Hotstuff) CompeteLeader() string {
	if p.smr.GetCurrentView() == 0 {
		return p.validatorsIP[0]
	}
	return p.smr.Election.GetLeader(p.smr.GetCurrentView())
}

func (p *Hotstuff) CompeteBlock(block IBlock) error {
	miner := p.CompeteLeader()
	logrus.Debugf("compete a leader(%s) address(%s) in round(%d)", miner, p.smr.GetAddress(), p.smr.GetCurrentView())
	if miner != p.smr.GetAddress() {
		return nil
	}
	proposal, err := p.smr.DoProposal(int64(block.GetHeight()), block.GetHash().Bytes(), p.validatorsIP)
	if err != nil {
		return err
	}

}
