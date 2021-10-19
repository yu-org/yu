package hotstuff

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/consensus/chained-hotstuff"
	. "github.com/yu-org/yu/tripod"
	"github.com/yu-org/yu/types"
)

type Hotstuff struct {
	meta         *TripodMeta
	validatorsIP []string

	smr *Smr

	msgChan chan []byte
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
		msgChan:      make(chan []byte, 10),
	}
}

func (h *Hotstuff) GetTripodMeta() *TripodMeta {
	return h.meta
}

func (h *Hotstuff) Name() string {
	return h.meta.Name()
}

func (h *Hotstuff) CheckTxn(txn *types.SignedTxn) error {
	return nil
}

func (h *Hotstuff) VerifyBlock(block *types.CompactBlock, env *ChainEnv) bool {
	return true
}

func (h *Hotstuff) InitChain(env *ChainEnv, _ *Land) error {
	chain := env.Chain
	gensisBlock := &types.CompactBlock{
		Header: &types.Header{},
	}
	return chain.SetGenesis(gensisBlock)
}

func (h *Hotstuff) StartBlock(block *types.CompactBlock, env *ChainEnv, land *Land) error {
	panic("implement me")
}

func (h *Hotstuff) EndBlock(block *types.CompactBlock, env *ChainEnv, land *Land) error {
	panic("implement me")
}

func (h *Hotstuff) FinalizeBlock(block *types.CompactBlock, env *ChainEnv, land *Land) error {
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

func (h *Hotstuff) CompeteLeader() string {
	if h.smr.GetCurrentView() == 0 {
		return h.validatorsIP[0]
	}
	return h.smr.Election.GetLeader(h.smr.GetCurrentView())
}

func (h *Hotstuff) CompeteBlock(block *types.CompactBlock) error {
	miner := h.CompeteLeader()
	logrus.Debugf("compete a leader(%s) address(%s) in round(%d)", miner, h.smr.GetAddress(), h.smr.GetCurrentView())
	if miner != h.smr.GetAddress() {
		return nil
	}
	proposal, err := h.smr.DoProposal(int64(block.Height), block.Hash.Bytes(), h.validatorsIP)
	if err != nil {
		return err
	}

}
