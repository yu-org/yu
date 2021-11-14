package hotstuff

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	"github.com/xuperchain/xupercore/lib/utils"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/consensus/chained-hotstuff"
	"github.com/yu-org/yu/context"
	"github.com/yu-org/yu/keypair"
	. "github.com/yu-org/yu/tripod"
	"github.com/yu-org/yu/types"
)

type Hotstuff struct {
	meta *TripodMeta
	// key: crypto address
	validators map[string]peer.ID
	myPubkey   keypair.PubKey
	myPrivKey  keypair.PrivKey

	smr *Smr

	env              *ChainEnv
	proposalDataChan chan []byte
	voteMsgChan      chan []byte
}

func NewHotstuff(myPubkey keypair.PubKey, myPrivkey keypair.PrivKey, validatorsMap map[string]string) *Hotstuff {
	meta := NewTripodMeta("hotstuff")

	q := InitQcTee()
	saftyrules := &DefaultSaftyRules{
		QcTree: q,
	}

	validatorsAddr := make([]string, 0)
	validators := make(map[string]peer.ID)
	for addr, ip := range validatorsMap {
		peerID, err := peer.Decode(ip)
		if err != nil {
			logrus.Fatalf("decode validatorIP(%s) error: %v", ip, err)
		}
		validators[addr] = peerID

		validatorsAddr = append(validatorsAddr, addr)
	}

	elec := NewSimpleElection(validatorsAddr)
	smr := NewSmr(myPubkey.String(), &DefaultPaceMaker{}, saftyrules, elec, q)

	h := &Hotstuff{
		meta:             meta,
		validators:       validators,
		myPubkey:         myPubkey,
		myPrivKey:        myPrivkey,
		smr:              smr,
		proposalDataChan: make(chan []byte, 10),
		voteMsgChan:      make(chan []byte, 10),
	}
	h.meta.SetP2pHandler(ProposeCode, h.handleRecvProposal).SetP2pHandler(VoteCode, h.handleRecvVoteMsg)
	h.meta.SetExec(h.JoinValidator, 10000).SetExec(h.QuitValidator, 100)
	return h
}

func (h *Hotstuff) ValidatorsP2pID() (peers []peer.ID) {
	for _, id := range h.validators {
		peers = append(peers, id)
	}
	return
}

func (h *Hotstuff) LocalAddress() string {
	return h.myPubkey.Address().String()
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

func (h *Hotstuff) SetChainEnv(env *ChainEnv) {
	h.env = env
}

func (h *Hotstuff) VerifyBlock(block *types.CompactBlock) bool {
	return true
}

func (h *Hotstuff) InitChain() error {
	chain := h.env.Chain
	gensisBlock := &types.CompactBlock{
		Header: &types.Header{},
	}
	return chain.SetGenesis(gensisBlock)
}

func (h *Hotstuff) StartBlock(block *types.CompactBlock) error {
	miner := h.CompeteLeader()
	logrus.Debugf("compete a leader(%s) address(%s) in round(%d)", miner, h.smr.GetAddress(), h.smr.GetCurrentView())
	if miner != h.smr.GetAddress() {
		return nil
	}

	return nil
}

func (h *Hotstuff) EndBlock(block *types.CompactBlock) error {
	chain := h.env.Chain
	pool := h.env.Pool

	err := h.env.Execute(block)
	if err != nil {
		return err
	}

	err = chain.AppendBlock(block)
	if err != nil {
		return err
	}

	logrus.Infof("append block(%d) (%s)", block.Height, block.Hash.String())

	h.env.SetCanRead(block.Hash)

	return pool.Reset()
}

func (h *Hotstuff) FinalizeBlock(block *types.CompactBlock) error {
	h.doPropose(int64(block.Height), block.Hash.Bytes())
	pNode := h.smr.BlockToProposalNode(block)
	err := h.smr.UpdateQcStatus(pNode)
	logrus.Debugf("Hotstuff::ProcessConfirmBlock::Now HighQC(%s) blockHash(%s) error: %v", utils.F(h.smr.GetHighQC().GetProposalId()), block.Hash.String(), err)
	return err
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
	return h.smr.Election.GetLeader(h.smr.GetCurrentView())
}

func (h *Hotstuff) JoinValidator(ctx *context.Context, block *types.CompactBlock) error {

	return nil
}

func (h *Hotstuff) QuitValidator(ctx *context.Context, block *types.CompactBlock) error {

	return nil
}
