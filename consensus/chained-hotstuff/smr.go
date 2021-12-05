// Copyright Xuperchain Authors
// link: https://github.com/xuperchain/xupercore

package chained_hotstuff

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	chainedBftPb "github.com/xuperchain/xupercore/kernel/consensus/base/driver/chained-bft/pb"
	"github.com/xuperchain/xupercore/lib/utils"
	"github.com/yu-org/yu/core/types"
	"sync"
	"time"
)

var (
	TooLowNewView      = errors.New("nextView is lower than local pacemaker's currentView.")
	P2PInternalErr     = errors.New("Internal err in p2p module.")
	TooLowNewProposal  = errors.New("Proposal is lower than local pacemaker's currentView.")
	EmptyHighQC        = errors.New("No valid highQC in qcTree.")
	SameProposalNotify = errors.New("Same proposal has been made.")
	JustifyVotesEmpty  = errors.New("justify qc's votes are empty.")
	EmptyTarget        = errors.New("Target parameter is empty.")
)

const (
	// DefaultNetMsgChanSize is the default size of network msg channel
	DefaultNetMsgChanSize = 1000
)

// smr 组装了三个模块: pacemaker、saftyrules和propose election
// smr有自己的存储即PendingTree
// 原本的ChainedBft(联结smr和本地账本，在preferredVote被确认后, 触发账本commit操作)
// 被替代成smr和上层bcs账本的·组合实现，以减少不必要的代码，考虑到chained-bft暂无扩展性
// 注意：本smr的round并不是强自增唯一的，不同节点可能产生相同round（考虑到上层账本的块可回滚）
type Smr struct {
	address string // 包含一个私钥生成的地址
	// subscribeList is the Subscriber list of the srm instance
	subscribeList *list.List

	pacemaker  IPacemaker
	saftyrules iSaftyRules
	Election   IProposerElection
	qcTree     *QCPendingTree
	// smr本地存储和外界账本存储的唯一关联，该字段标识了账本状态，
	// 但此处并不直接使用ledger handler作为变量，旨在结偶smr存储和本地账本存储
	// smr存储应该仅仅是账本区块头存储的很小的子集
	ledgerState int64

	// map[proposalId]bool
	localProposal *sync.Map
	// votes of QC in mem, key: voteId, value: []*QuorumCertSign
	qcVoteMsgs *sync.Map
}

func NewSmr(address string, pacemaker IPacemaker, saftyrules iSaftyRules,
	elec IProposerElection, qcTree *QCPendingTree) *Smr {
	s := &Smr{
		address:       address,
		subscribeList: list.New(),
		pacemaker:     pacemaker,
		saftyrules:    saftyrules,
		Election:      elec,
		qcTree:        qcTree,
		localProposal: &sync.Map{},
		qcVoteMsgs:    &sync.Map{},
	}
	s.localProposal.Store(utils.F(qcTree.Root.In.GetProposalId()), true)
	return s
}

func (s *Smr) DoPropose(viewNumber int64, proposalID []byte, validatesIpInfo []peer.ID) (*chainedBftPb.ProposalMsg, error) {
	// ATTENTION::TODO:: 由于本次设计面向的是viewNumber可能重复的BFT，因此账本回滚后高度会相同，在此用LockedQC高度为标记
	if validatesIpInfo == nil {
		return nil, EmptyTarget
	}
	if s.pacemaker.GetCurrentView() != s.qcTree.Genesis.In.GetProposalView()+1 &&
		s.qcTree.GetLockedQC() != nil && s.pacemaker.GetCurrentView() < s.qcTree.GetLockedQC().In.GetProposalView() {
		logrus.Debug("smr::ProcessProposal error: ", TooLowNewProposal, "pacemaker view", s.pacemaker.GetCurrentView(), "lockQC view",
			s.qcTree.GetLockedQC().In.GetProposalView())
		return nil, TooLowNewProposal
	}
	if s.qcTree.GetHighQC() == nil {
		logrus.Error("smr::ProcessProposal empty HighQC error")
		return nil, EmptyHighQC
	}
	if _, ok := s.localProposal.Load(utils.F(proposalID)); ok {
		return nil, SameProposalNotify
	}
	parentQuorumCert, err := s.reloadJustifyQC()
	if err != nil {
		logrus.Error("smr::ProcessProposal reloadJustifyQC error", "err", err)
		return nil, err
	}
	parentQuorumCertBytes, err := json.Marshal(parentQuorumCert)
	if err != nil {
		return nil, err
	}
	return &chainedBftPb.ProposalMsg{
		ProposalView: viewNumber,
		ProposalId:   proposalID,
		Timestamp:    time.Now().UnixNano(),
		JustifyQC:    parentQuorumCertBytes,
	}, nil
}

func (s *Smr) reloadJustifyQC() (*QuorumCert, error) {
	highQC := s.qcTree.GetHighQC()
	v := &VoteInfo{
		ProposalView: highQC.In.GetProposalView(),
		ProposalId:   highQC.In.GetProposalId(),
	}
	// 第一次proposal，highQC==rootQC==genesisQC
	if bytes.Equal(s.qcTree.Genesis.In.GetProposalId(), highQC.In.GetProposalId()) {
		return &QuorumCert{VoteInfo: v}, nil
	}
	// 查看qcTree是否包含当前可以commit的Id
	var commitId []byte
	if s.qcTree.GetCommitQC() != nil {
		commitId = s.qcTree.GetCommitQC().In.GetProposalId()
	}
	// 根据qcTree生成一个parentQC
	parentQuorumCert := &QuorumCert{
		VoteInfo: v,
		LedgerCommitInfo: &LedgerCommitInfo{
			CommitStateId: commitId,
		},
	}
	// 上一个view的votes
	value, ok := s.qcVoteMsgs.Load(utils.F(v.ProposalId))
	if !ok {
		return nil, JustifyVotesEmpty
	}
	signs, ok := value.([]*chainedBftPb.QuorumCertSign)
	if ok {
		parentQuorumCert.SignInfos = signs
	}
	return parentQuorumCert, nil
}

func (s *Smr) GetVoteAndParentQC(msg *chainedBftPb.ProposalMsg) (*VoteInfo, *QuorumCert, error) {
	if _, ok := s.localProposal.LoadOrStore(utils.F(msg.GetProposalId()), true); ok {
		return nil, nil, nil
	}
	logrus.Debug("smr::handleReceivedProposal::received a proposal",
		"newView", msg.GetProposalView(), "newProposalId", utils.F(msg.GetProposalId()))
	parentQCBytes := msg.GetJustifyQC()
	parentQC := &QuorumCert{}
	err := json.Unmarshal(parentQCBytes, parentQC)
	if err != nil {
		return nil, nil, err
	}
	newVote := &VoteInfo{
		ProposalId:   msg.GetProposalId(),
		ProposalView: msg.GetProposalView(),
		ParentId:     parentQC.GetProposalId(),
		ParentView:   parentQC.GetProposalView(),
	}
	return newVote, parentQC, nil
}

// CheckViewAndRound 和 HandleReceivedProposal 阶段在收到一个ProposalMsg后触发，与LibraBFT的process_proposal阶段类似
// 该阶段分两个角色，一个是认为自己是currentRound的Leader，一个是Replica
// 0. 查看ProposalMsg消息的合法性
// 1. 检查新的view是否符合账本状态要求
// 2. 比较本地pacemaker是否需要AdvanceRound
func (s *Smr) CheckViewAndRound(msg *chainedBftPb.ProposalMsg, newVote *VoteInfo, parentQC *QuorumCert) (needSendMsg bool, err error) {
	isFirstJustify := bytes.Equal(s.qcTree.Genesis.In.GetProposalId(), parentQC.GetProposalId())
	if !isFirstJustify {
		if err = s.saftyrules.CheckProposal(&QuorumCert{
			VoteInfo:  newVote,
			SignInfos: []*chainedBftPb.QuorumCertSign{msg.GetSign()},
		}, parentQC, s.Election.GetValidators(parentQC.GetProposalView())); err != nil {
			logrus.Debug("smr::handleReceivedProposal::CheckProposal error", "error", err,
				"parentView", parentQC.GetProposalView(), "parentId", utils.F(parentQC.GetProposalId()))
			return
		}
	}
	/*
		if !bytes.Equal(parentQC.GetProposalId(), s.qcTree.GetHighQC().In.GetProposalId()) {
			return	// TODO: 新的proposal需要严格保证在HighQC下面，否则不参与投票
		}
	*/
	// 1.检查账本状态和收到新round是否符合要求
	if s.ledgerState+3 < newVote.ProposalView {
		logrus.Error("smr::handleReceivedProposal::local ledger hasn't been updated.", "LedgerState", s.ledgerState, "ProposalView", newVote.ProposalView)
		return
	}
	// 2.本地pacemaker试图更新currentView, 并返回一个是否需要将新消息通知该轮Leader, 是该轮不是下轮！主要解决P2PIP端口不能通知Loop的问题
	needSendMsg, _ = s.pacemaker.AdvanceView(parentQC)
	logrus.Debug("smr::handleReceivedProposal::pacemaker update", "view", s.pacemaker.GetCurrentView())
	// 通知current Leader
	return
}

// 3. 检查qcTree是否需要更新CommitQC
// 4. 查看收到的view是否符合要求
// 5. 向本地PendingTree插入该QC，即更新QC
// 6. 发送一个vote消息给下一个Leader
// 注意：该过程删除了当前round的leader是否符合计算，将该步骤后置到上层共识CheckMinerMatch，原因：需要支持上层基于时间调度而不是基于round调度，减小耦合
func (s *Smr) HandleRecvProposal(msg *chainedBftPb.ProposalMsg, newVote *VoteInfo, parentQC *QuorumCert) (*chainedBftPb.VoteMsg, string, error) {

	// 3.本地safetyrules更新, 如有可以commit的QC，执行commit操作并更新本地rootQC
	if parentQC.LedgerCommitInfo != nil && parentQC.LedgerCommitInfo.CommitStateId != nil &&
		s.saftyrules.UpdatePreferredRound(parentQC.GetProposalView()) {
		s.qcTree.updateCommit(parentQC.GetProposalId())
	}
	// 4.查看收到的view是否符合要求, 此处接受孤儿节点
	if !s.saftyrules.CheckPacemaker(msg.GetProposalView(), s.pacemaker.GetCurrentView()) {
		logrus.Error("smr::handleReceivedProposal::error", "error", TooLowNewProposal, "local want", s.pacemaker.GetCurrentView(),
			"proposal have", msg.GetProposalView())
		return nil, "", nil
	}

	// 注意：删除此处的验证收到的proposal是否符合local计算，在本账本状态中后置到上层共识CheckMinerMatch
	// 根据本地saftyrules返回是否 需要发送voteMsg给下一个Leader
	if !s.saftyrules.VoteProposal(msg.GetProposalId(), msg.GetProposalView(), parentQC) {
		logrus.Error("smr::handleReceivedProposal::VoteProposal fail", "view", msg.GetProposalView(), "proposalId", msg.GetProposalId())
		return nil, "", nil
	}

	// 这个newVoteId表示的是本地最新一次vote的id，生成voteInfo的hash，标识vote消息
	newLedgerInfo := &LedgerCommitInfo{
		VoteInfoHash: msg.GetProposalId(),
	}
	newNode := &ProposalNode{
		In: &QuorumCert{
			VoteInfo:         newVote,
			LedgerCommitInfo: newLedgerInfo,
		},
	}
	// 5.与proposal.ParentId相比，更新本地qcTree，insert新节点, 包括更新CommitQC等等
	err := s.qcTree.updateQcStatus(newNode)
	if err != nil {
		return nil, "", err
	}
	logrus.Debug("smr::handleReceivedProposal::pacemaker changed", "round", s.pacemaker.GetCurrentView())
	// 6.发送一个vote消息给下一个Leader
	nextLeader := s.Election.GetLeader(s.pacemaker.GetCurrentView() + 1)
	if nextLeader == "" {
		logrus.Debug("smr::handleReceivedProposal::empty next leader", "next round", s.pacemaker.GetCurrentView()+1)
		return nil, "", nil
	}
	// 若为自己直接先返回
	if nextLeader == s.address {
		return nil, "", nil
	}
	voteBytes, err := json.Marshal(newVote)
	if err != nil {
		return nil, "", err
	}
	ledgerBytes, err := json.Marshal(newLedgerInfo)
	if err != nil {
		return nil, "", err
	}
	return &chainedBftPb.VoteMsg{
		VoteInfo:         voteBytes,
		LedgerCommitInfo: ledgerBytes,
		Signature:        []*chainedBftPb.QuorumCertSign{ /* nextSign */ },
	}, nextLeader, nil
}

// handleReceivedVoteMsg 当前Leader在发送一个proposal消息之后，由下一Leader等待周围replica的投票，收集vote消息
// 当收到2f+1个vote消息之后，本地pacemaker调用AdvanceView，并更新highQC
// 该方法针对Leader而言
// 如果超过2f+1，则返回 true, nil
func (s *Smr) HandleRecvVoteMsg(msg *chainedBftPb.VoteMsg) (bool, error) {
	voteQC, err := VoteMsgToQC(msg)
	if err != nil {
		logrus.Error("smr::handleReceivedVoteMsg VoteMsgToQC error", "error", err)
		return false, err
	}
	// 检查voteInfoHash是否正确
	if err := s.saftyrules.CheckVote(voteQC, s.Election.GetValidators(voteQC.GetProposalView())); err != nil {
		logrus.Error("smr::handleReceivedVoteMsg CheckVote error", "error", err, "msg", utils.F(voteQC.GetProposalId()))
		return false, err
	}
	logrus.Debug("smr::handleReceivedVoteMsg::receive vote", "voteId", utils.F(voteQC.GetProposalId()), "voteView", voteQC.GetProposalView(), "from", voteQC.SignInfos[0].Address)

	// 若vote先于proposal到达，则直接丢弃票数
	if _, ok := s.localProposal.Load(utils.F(voteQC.GetProposalId())); !ok {
		logrus.Debug("smr::handleReceivedVoteMsg::haven't received the related proposal msg, drop it.")
		return false, EmptyTarget
	}
	if node := s.qcTree.DFSQueryNode(voteQC.GetProposalId()); node == nil {
		logrus.Debug("smr::handleReceivedVoteMsg::haven't finish proposal process, drop it.")
		return false, EmptyTarget
	}

	// 存入本地voteInfo内存，查看签名数量是否超过2f+1
	var VoteLen int
	// 注意隐式，若!ok则证明签名数量为1，此时不可能超过2f+1
	v, ok := s.qcVoteMsgs.LoadOrStore(utils.F(voteQC.GetProposalId()), voteQC.SignInfos)
	// 若ok=false，则仅store一个vote签名
	VoteLen = 1
	if ok {
		signs, _ := v.([]*chainedBftPb.QuorumCertSign)
		stored := false
		for _, sign := range signs {
			// 自己给自己投票将自动忽略
			if sign.Address == voteQC.SignInfos[0].Address || voteQC.SignInfos[0].Address == s.address {
				stored = true
			}
		}
		if !stored {
			signs = append(signs, voteQC.SignInfos[0])
			s.qcVoteMsgs.Store(utils.F(voteQC.GetProposalId()), signs)
		}
		VoteLen = len(signs)
	}
	// 查看签名数量是否达到2f+1, 需要获取justify对应的validators
	if !s.saftyrules.CalVotesThreshold(VoteLen, len(s.Election.GetValidators(voteQC.GetProposalView()))) {
		return false, nil
	}

	// 更新本地pacemaker AdvanceRound
	s.pacemaker.AdvanceView(voteQC)
	logrus.Debug("smr::handleReceivedVoteMsg::FULL VOTES!", "pacemaker view", s.pacemaker.GetCurrentView())
	// 更新HighQC
	s.qcTree.updateHighQC(voteQC.GetProposalId())
	return true, nil
}

func (s *Smr) GetCurrentView() int64 {
	return s.pacemaker.GetCurrentView()
}

func (s *Smr) GetAddress() string {
	return s.address
}

func (s *Smr) GetSaftyRules() iSaftyRules {
	return s.saftyrules
}

func (s *Smr) GetPacemaker() IPacemaker {
	return s.pacemaker
}

func (s *Smr) GetHighQC() IQuorumCert {
	return s.qcTree.GetHighQC().In
}

// GetCompleteHighQC 本地qcTree不带签名，因此smr需要重新组装完整的QC
func (s *Smr) GetCompleteHighQC() IQuorumCert {
	raw := s.qcTree.GetHighQC().In
	renew := &QuorumCert{
		VoteInfo: &VoteInfo{
			ProposalId:   raw.GetProposalId(),
			ProposalView: raw.GetProposalView(),
		},
	}
	if raw.GetParentProposalId() != nil {
		renew.VoteInfo.ParentId = raw.GetParentProposalId()
		renew.VoteInfo.ParentView = raw.GetProposalView()
	}
	signInfo, ok := s.qcVoteMsgs.Load(utils.F(raw.GetProposalId()))
	if !ok {
		return renew
	}
	signs, _ := signInfo.([]*chainedBftPb.QuorumCertSign)
	renew.SignInfos = signs
	return renew
}

func (s *Smr) GetGenericQC() IQuorumCert {
	if s.qcTree.GetGenericQC() == nil {
		return nil
	}
	return s.qcTree.GetGenericQC().In
}

func (s *Smr) UpdateQcStatus(node *ProposalNode) error {
	if node == nil {
		return EmptyTarget
	}
	// 更新ledgerStatus
	if node.In.GetProposalView() > s.ledgerState {
		s.ledgerState = node.In.GetProposalView()
	}
	return s.qcTree.updateQcStatus(node)
}

func (s *Smr) EnforceUpdateHighQC(inProposalId []byte) error {
	return s.qcTree.enforceUpdateHighQC(inProposalId)
}

func (s *Smr) BlockToProposalNode(block *types.CompactBlock) *ProposalNode {
	blockHash := block.Hash
	node := s.qcTree.DFSQueryNode(blockHash.Bytes())
	if node != nil {
		return node
	}
	height := int64(block.Height)
	return &ProposalNode{
		In: &QuorumCert{
			VoteInfo: &VoteInfo{
				ProposalId:   blockHash.Bytes(),
				ProposalView: height,
				ParentId:     block.PrevHash.Bytes(),
				ParentView:   height - 1,
			},
		},
	}
}

func VoteMsgToQC(msg *chainedBftPb.VoteMsg) (*QuorumCert, error) {
	voteInfo := &VoteInfo{}
	if err := json.Unmarshal(msg.VoteInfo, voteInfo); err != nil {
		return nil, err
	}
	ledgerCommitInfo := &LedgerCommitInfo{}
	if err := json.Unmarshal(msg.LedgerCommitInfo, ledgerCommitInfo); err != nil {
		return nil, err
	}
	return &QuorumCert{
		VoteInfo:         voteInfo,
		LedgerCommitInfo: ledgerCommitInfo,
		SignInfos:        msg.GetSignature(),
	}, nil
}
