package hotstuff

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	"github.com/xuperchain/xupercore/lib/utils"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/consensus/chained-hotstuff"
	"github.com/yu-org/yu/context"
	"github.com/yu-org/yu/keypair"
	. "github.com/yu-org/yu/tripod"
	. "github.com/yu-org/yu/types"
)

type Hotstuff struct {
	meta *TripodMeta
	// key: crypto address
	validators map[string]peer.ID
	myPubkey   keypair.PubKey
	myPrivKey  keypair.PrivKey

	smr *Smr

	env       *ChainEnv
	blockChan chan *Block
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
		meta:       meta,
		validators: validators,
		myPubkey:   myPubkey,
		myPrivKey:  myPrivkey,
		smr:        smr,
		blockChan:  make(chan *Block, 10),
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

func (h *Hotstuff) CheckTxn(txn *SignedTxn) error {
	return nil
}

func (h *Hotstuff) SetChainEnv(env *ChainEnv) {
	h.env = env
}

func (h *Hotstuff) VerifyBlock(block *CompactBlock) bool {
	return true
}

func (h *Hotstuff) InitChain() error {
	chain := h.env.Chain
	gensisBlock := &CompactBlock{
		Header: &Header{},
	}
	err := chain.SetGenesis(gensisBlock)
	if err != nil {
		return err
	}
	go func() {
		for {
			msg, err := h.env.P2pNetwork.SubP2P(StartBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P error: ", err)
				continue
			}
			block, err := DecodeBlock(msg)
			if err != nil {
				logrus.Error("decode block from p2p error: ", err)
				continue
			}
			h.blockChan <- block
		}
	}()
	return nil
}

func (h *Hotstuff) StartBlock(block *CompactBlock) error {
	miner := h.CompeteLeader()
	logrus.Debugf("compete a leader(%s) address(%s) in round(%d)", miner, h.smr.GetAddress(), h.smr.GetCurrentView())
	if miner != h.smr.GetAddress() {
		h.useP2pBlock(block)
		return nil
	}

	// TODO: pack txns from txpool
	txns, err := h.env.Pool.Pack(3000)
	if err != nil {
		return err
	}
	hashes := FromArray(txns...).Hashes()
	block.TxnsHashes = hashes

	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		return err
	}
	block.TxnRoot = txnRoot

	byt, _ := block.Encode()
	block.Hash = BytesToHash(Sha256(byt))

	h.env.StartBlock(block.Hash)
	err = h.env.Base.SetTxns(block.Hash, txns)
	if err != nil {
		return err
	}

	rawBlock := &Block{
		CompactBlock: block,
		Txns:         txns,
	}

	rawBlockByt, err := rawBlock.Encode()
	if err != nil {
		return err
	}

	return h.env.P2pNetwork.PubP2P(StartBlockTopic, rawBlockByt)
}

func (h *Hotstuff) EndBlock(block *CompactBlock) error {
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

func (h *Hotstuff) FinalizeBlock(block *CompactBlock) error {
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

func (h *Hotstuff) useP2pBlock(localBlock *CompactBlock) {
	p2pBlock := <-h.blockChan
	ok := h.VerifyBlock(p2pBlock.CompactBlock)
	if !ok {
		logrus.Warnf("block(%s) verify failed", p2pBlock.Hash.String())
		return
	}
	localBlock.CopyFrom(p2pBlock.CompactBlock)
	err := h.env.Base.SetTxns(localBlock.Hash, p2pBlock.Txns)
	if err != nil {
		logrus.Errorf("set txns of p2p-block(%s) into base error: %v", p2pBlock.Hash.String(), err)
		return
	}
	h.env.StartBlock(localBlock.Hash)
	err = h.env.Pool.RemoveTxns(localBlock.TxnsHashes)
	if err != nil {
		logrus.Error("clear txpool error: ", err)
	}
}

func (h *Hotstuff) JoinValidator(ctx *context.Context, block *CompactBlock) error {

	return nil
}

func (h *Hotstuff) QuitValidator(ctx *context.Context, block *CompactBlock) error {

	return nil
}
