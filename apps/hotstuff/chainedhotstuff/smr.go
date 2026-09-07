package chainedhotstuff

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/sirupsen/logrus"
	pb "github.com/yu-org/yu/hotstuff/chainedhotstuff/proto"
	"github.com/yu-org/yu/hotstuff/chainedhotstuff/storage"

	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/hotstuff/chainedhotstuff/crypto"
)

type IBlock interface {
	GetProposer() []byte
	GetHeight() int64
	GetBlockid() []byte
	GetConsensusStorage() ([]byte, error)
	GetTimestamp() int64
	SetItem(item string, value interface{}) error
	MakeBlockId() ([]byte, error)
	GetPreHash() []byte
	GetNextHash() []byte
	GetPublicKey() string
	GetSign() []byte
	GetTxIDs() []string
	GetInTrunk() bool
}

func VerifyChecksum(msg *pb.XuperMessage) bool {
	return crc32.ChecksumIEEE(msg.GetData().GetMsgInfo()) == msg.GetHeader().GetDataCheckSum()
}

func Unmarshal(msg *pb.Message, message proto.Message) error {
	if !VerifyChecksum(msg) {
		return errors.New("verify checksum error")
	}

	data, err := Decompress(msg)
	if err != nil {
		return errors.New("decompress error")
	}

	err = proto.Unmarshal(data, message)
	if err != nil {
		return errors.New("message unmarshal error")
	}

	return nil
}

type MessageOption func(*pb.Message)

func GenPseudoUniqId() uint64 {
	nano := time.Now().UnixNano()

	randNum1 := rand.Int63()
	randNum2 := rand.Int63()
	shift1 := rand.Intn(16) + 2
	shift2 := rand.Intn(8) + 1

	uId := ((randNum1 >> uint(shift1)) + (randNum2 >> uint(shift2)) + (nano >> 1)) &
		0x1FFFFFFFFFFFFF
	return uint64(uId)

}

func GenLogId() string {
	return fmt.Sprintf("%d_%d", time.Now().Unix(), GenPseudoUniqId())
}

func NewMessage(ty pb.Message_MsgType, message proto.Message, opts ...MessageOption) pb.Message {
	msg := pb.Message{
		Header: &pb.Message_MsgHeader{
			Version:        "3.0.0",
			Bcname:         "yu",
			Logid:          GenLogId(),
			Type:           ty,
			EnableCompress: false,
			ErrorType:      pb.Message_NONE,
		},
		Data: message.Message_MsgData{},
	}

	if message != nil {
		data, _ := proto.Marshal(message)
		msg.Data.MsgInfo = data
	}

	for _, op := range opts {
		op(msg)
	}
	Compress(msg)
	msg.Header.DataCheckSum = Checksum(msg)
	return msg
}

func WithBCName(bcname string) MessageOption {
	return func(msg *pb.Message) {
		msg.Header.Bcname = bcname
	}
}

type MarkPoint struct {
	tag   string
	delta float64
}

type Timer struct {
	bornTime   int64
	latestTime int64
	points     []*MarkPoint
}

func NewTimer() *Timer {
	now := time.Now().UnixNano()
	return &Timer{
		bornTime:   now,
		latestTime: now,
	}
}

var (
	ErrTooLowNewView      = errors.New("nextView is lower than local pacemaker's currentView")
	ErrP2PInternalErr     = errors.New("internal err in p2p module")
	ErrTooLowNewProposal  = errors.New("proposal is lower than local pacemaker's currentView")
	ErrEmptyHighQC        = errors.New("no valid highQC in qcTree")
	ErrSameProposalNotify = errors.New("same proposal has been made")
	ErrJustifyVotesEmpty  = errors.New("justify qc's votes are empty")
	ErrEmptyTarget        = errors.New("target parameter is empty")
	ErrRegisterErr        = errors.New("register to p2p error")
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
	bcName  string
	log     loglogrus.er
	address string // 包含一个私钥生成的地址
	// smr定义了自己需要的P2P消息类型
	// p2pMsgChan is the msg channel registered to network
	//p2pMsgChan chan *xuperp2p.XuperMessage
	// p2p interface
	p2p *LibP2P
	// cBFTCrypto 封装了本BFT需要的加密相关的接口和方法
	cryptoClient *crypto.Crypto

	// quitCh stop channel
	quitCh chan bool

	pacemaker  PacemakerInterface
	saftyrules saftyRulesInterface
	election   IProposerElection
	qcTree     *storage.QCPendingTree
	// smr本地存储和外界账本存储的唯一关联，该字段标识了账本状态，
	// 但此处并不直接使用ledger handler作为变量，旨在结偶smr存储和本地账本存储
	// smr存储应该仅仅是账本区块头存储的很小的子集
	ledgerState int64

	// map[proposalId]int64
	localProposal *sync.Map
	// votes of QC in mem, key: voteId, value: []*QuorumCertSign
	qcVoteMsgs *sync.Map

	// 该锁保护状态机处理msg或者bcs层操作过程，防止状态机get/set时由于bcs操作和msg处理并发导致的脏读脏写
	mtx sync.Mutex
}

func NewSmr(bcName, address string, log loglogrus.er, p2p *LibP2P, cryptoClient *crypto.Crypto, pacemaker PacemakerInterface,
	saftyrules saftyRulesInterface, election IProposerElection, qcTree *storage.QCPendingTree) *Smr {
	s := &Smr{
		bcName:  bcName,
		log:     log,
		address: address,
		//p2pMsgChan:    make(chan *xuperp2p.XuperMessage, DefaultNetMsgChanSize),
		p2p:           p2p,
		cryptoClient:  cryptoClient,
		quitCh:        make(chan bool, 1),
		pacemaker:     pacemaker,
		saftyrules:    saftyrules,
		election:      election,
		qcTree:        qcTree,
		localProposal: &sync.Map{},
		qcVoteMsgs:    &sync.Map{},
	}
	// smr初始值装载
	s.localProposal.Store(fmt.Sprintf("%x", qcTree.GetRootQC().In.GetProposalId()), 0)
	if qcTree.GetHighQC() != nil {
		s.ledgerState = int64(qcTree.GetHighQC().In.GetProposalView())
	} else if qcTree.GetGenericQC() != nil {
		s.ledgerState = int64(qcTree.GetGenericQC().In.GetProposalView())
	} else {
		s.ledgerState = int64(qcTree.GetRootQC().In.GetProposalView())
	}
	return s
}

func (s *Smr) LoadVotes(proposalId []byte, signs []*pb.QuorumCertSignature) {
	if signs != nil {
		s.qcVoteMsgs.Store(fmt.Sprintf("%x", proposalId), signs)
	}
}

func MsgTypeToString(msgType pb.MsgType) string {
	switch msgType {
	case MsgType_NEW_VIEW:
		return "MSG_NEW_VIEW"
	case MsgType_NEW_PROPOSAL:
		return "MSG_NEW_PROPOSAL"
	case MsgType_VOTE:
		return "MSG_VOTE"
	}

	return nil
}

func (s *Smr) RegisterToNetwork() error {
	sub1, err := s.p2p.SubP2P(MsgTypeToString(pb.MsgType_NEW_VIEW))
	if err != nil {
		return err
	}

	sub2, err := s.p2p.SubP2P(MsgTypeToString(pb.MsgType_NEW_PROPOSAL))
	if err != nil {
		return err
	}

	sub3, err := s.p2p.SubP2P(MsgTypeToString(pb.MsgType_NEW_PROPOSAL))
	if err != nil {
		return err
	}
}

// Start used to start smr instance and process msg
func (s *Smr) Start() {
	s.RegisterToNetwork()
	go func() {
		for {
			select {
			case msg := <-s.p2pMsgChan:
				s.handleReceivedMsg(msg)
			case <-s.quitCh:
				return
			}
		}
	}()
}

// stop used to stop smr instance
func (s *Smr) Stop() {
	s.quitCh <- true
	s.UnRegisterToNetwork()
}

// GetRootQC 查询状态树的Root节点，Root节点已经被账本commit
func (s *Smr) GetRootQC() storage.QuorumCertInterface {
	return s.qcTree.GetRootQC().In
}

func (s *Smr) GetCurrentView() int64 {
	return s.pacemaker.GetCurrentView()
}

func (s *Smr) GetAddress() string {
	return s.address
}

func (s *Smr) CheckProposal(block IBlock, justify storage.QuorumCertInterface, validators []string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	pNode := s.blockToProposalNode(block)
	return s.saftyrules.CheckProposal(pNode.In, justify, validators)
}

func (s *Smr) KeepUpWithBlock(block IBlock, justify storage.QuorumCertInterface, validators []string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.updateJustifyQcStatus(justify)
	if validators != nil {
		err := s.ProcessProposal(block.GetHeight(), block.GetBlockid(), block.GetPreHash(), validators)
		if err != nil && err != ErrSameProposalNotify && err != ErrTooLowNewProposal {
			return err
		}
	}
	// 在不在候选人节点中，都直接调用smr生成新的qc树，矿工调用避免了proposal消息后于vote消息
	pNode := s.blockToProposalNode(block)
	err := s.updateQcStatus(pNode)
	if err != nil {
		return err
	}
	s.qcTree.UpdateCommit(block.GetPreHash())
	s.pacemaker.AdvanceView(justify)
	logrus.Debug("consensus:smr:KeepUpWithBlock: current parameters: ", "highQC", fmt.Sprintf("%x", s.getHighQC().GetProposalId()), "blockId", fmt.Sprintf("%x", block.GetBlockid()),
		"pacemaker view", s.pacemaker.GetCurrentView(), "QCTree Root", fmt.Sprintf("%x", s.qcTree.GetRootQC().In.GetProposalId()))
	return nil
}

func (s *Smr) ResetProposerStatus(tipBlock IBlock,
	queryBlockFunc func(blkId []byte) (IBlock, error),
	validators []string) (bool, storage.QuorumCertInterface, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if bytes.Equal(s.getHighQC().GetProposalId(), tipBlock.GetBlockid()) &&
		s.validNewHighQC(tipBlock.GetBlockid(), validators) {
		// 此处需要获取带签名的完整Justify
		return false, s.getCompleteHighQC(), nil
	}

	// 从当前TipBlock开始往前追溯，交给smr根据状态进行回滚。
	// 在本地状态树上找到指代TipBlock的QC，若找不到，则在状态树上找和TipBlock同一分支上的最近值
	var qc storage.QuorumCertInterface
	targetId := tipBlock.GetBlockid()
	for {
		block, err := queryBlockFunc(targetId)
		if err != nil {
			logrus.Error("consensus:smr:ResetProposerStatus: queryBlockFunc error.", "error", err)
			return false, nil, ErrEmptyTarget
		}
		// 至多回滚到root节点
		if block.GetHeight() <= s.GetRootQC().GetProposalView() {
			logrus.Warn("consensus:smr:ResetProposerStatus: set root qc.", "root", fmt.Sprintf("%x", s.GetRootQC().GetProposalId()), "root height", s.GetRootQC().GetProposalView(),
				"block", fmt.Sprintf("%x", block.GetBlockid()), "block height", block.GetHeight())
			qc = s.GetRootQC()
			break
		}
		// 查找目标Id是否挂在状态树上，若否，则从target网上查找知道状态树里有
		node := s.qcTree.DFSQueryNode(block.GetBlockid())
		if node == nil {
			targetId = block.GetPreHash()
			continue
		}
		// node在状态树上找到之后，以此为起点(包括当前点)，继续向上查找，知道找到符合全名数量要求的QC，该QC可强制转化为新的HighQC
		wantProposers := s.election.GetValidators(block.GetHeight())
		if wantProposers == nil {
			logrus.Error("consensus:smr:ResetProposerStatus: election error.")
			return false, nil, ErrEmptyTarget
		}
		if !s.validNewHighQC(node.In.GetProposalId(), wantProposers) {
			logrus.Warn("consensus:smr:ResetProposerStatus: target not ready", "target", fmt.Sprintf("%x", node.In.GetProposalId()), "wantProposers", wantProposers, "height", node.In.GetProposalView())
			targetId = block.GetPreHash()
			continue
		}
		qc = node.In
		break
	}
	if qc == nil {
		return false, nil, ErrEmptyHighQC
	}
	ok, err := s.enforceUpdateHighQC(qc.GetProposalId())
	if err != nil {
		logrus.Error("consensus:smr:ResetProposerStatus: EnforceUpdateHighQC error.", "error", err)
		return false, nil, err
	}
	if ok {
		logrus.Debug("consensus:smr:ResetProposerStatus: EnforceUpdateHighQC success.", "target", fmt.Sprintf("%x", qc.GetProposalId()), "height", qc.GetProposalView())
	}
	// 此处需要获取带签名的完整Justify, 此时HighQC已经更新
	return true, s.getCompleteHighQC(), nil
}

// handleReceivedMsg used to process msg received from network
func (s *Smr) handleReceivedMsg(msg *pb.Message) error {
	// filter msg from other chain
	if msg.GetHeader().GetBcname() != s.bcName {
		return nil
	}
	switch msg.GetHeader().GetType() {
	case pb.Message_NEW_PROPOSAL_MSG:
		s.handleReceivedProposal(msg)
	case pb.Message_VOTE_MSG:
		s.handleReceivedVoteMsg(msg)
	default:
		logrus.Error("smr::handleReceivedMsg receive unknow type msg", "type", msg.GetHeader().GetType())
		return nil
	}
	return nil
}

// UpdateJustifyQcStatus 用于支持可回滚的账本，生成相同高度的块
// 为了支持生成相同round的块，需要拿到justify的full votes，因此需要在上层账本收到新块时调用，在CheckMinerMatch后
// 注意：为了支持回滚操作，必须调用该函数
func (s *Smr) updateJustifyQcStatus(justify storage.QuorumCertInterface) {
	if justify == nil {
		return
	}
	v, ok := s.qcVoteMsgs.Load(fmt.Sprintf("%x", justify.GetProposalId()))
	var signs []*pb.QuorumCertSignature
	if ok {
		signs, _ = v.([]*pb.QuorumCertSignature)
	}
	justifySigns := justify.GetSignsInfo()
	if justifySigns == nil {
		return
	}
	signs = appendSigns(signs, justifySigns)
	s.qcVoteMsgs.Store(fmt.Sprintf("%x", justify.GetProposalId()), signs)
	// 根据justify check情况更新本地HighQC, 注意：由于CheckMinerMatch已经检查justify签名
	s.qcTree.UpdateHighQC(justify.GetProposalId())
}

// UpdateQcStatus 除了更新本地smr的QC之外，还更新了smr的和账本相关的状态，以此区别于smr receive proposal时的updateQcStatus
func (s *Smr) updateQcStatus(node *storage.ProposalNode) error {
	if node == nil {
		return ErrEmptyTarget
	}
	// 更新ledgerStatus
	if node.In.GetProposalView() > s.ledgerState {
		s.ledgerState = node.In.GetProposalView()
	}
	return s.qcTree.UpdateQcStatus(node)
}

// ProcessProposal 即Chained-HotStuff的NewView阶段，LibraBFT的process_proposal阶段
// 对于一个认为自己当前是Leader的节点，它试图生成一个新的提案，即一个新的QC，并广播
// 本节点产生一个Proposal，该proposal包含一个最新的round, 最新的proposalId，一个parentQC，并将该消息组合成一个ProposalMsg消息给所有节点
// 全部完成后leader更新本地localProposal
func (s *Smr) ProcessProposal(viewNumber int64, proposalID []byte, parentID []byte, validatesIpInfo []string) error {
	// ATTENTION::TODO:: 由于本次设计面向的是viewNumber可能重复的BFT，因此账本回滚后高度会相同，在此用LockedQC高度为标记
	if validatesIpInfo == nil {
		return ErrEmptyTarget
	}
	if s.pacemaker.GetCurrentView() != s.qcTree.GetGenesisQC().In.GetProposalView()+1 &&
		s.qcTree.GetLockedQC() != nil && s.pacemaker.GetCurrentView() < s.qcTree.GetLockedQC().In.GetProposalView() {
		logrus.Error("smr::ProcessProposal error", "error", ErrTooLowNewProposal, "pacemaker view", s.pacemaker.GetCurrentView(), "lockQC view",
			s.qcTree.GetLockedQC().In.GetProposalView())
		return ErrTooLowNewProposal
	}
	if s.getHighQC() == nil {
		logrus.Error("smr::ProcessProposal empty HighQC error")
		return ErrEmptyHighQC
	}
	if _, ok := s.localProposal.Load(fmt.Sprintf("%x", proposalID)); ok {
		return ErrSameProposalNotify
	}
	// Libra-BFT中的parentQC为本地HighQC，但由于本系统支持回滚，故HighQC有可能在新QC生成时变更，否则会导致QC序错误
	// 故本系统的parentQC必须提前指定，不能是highQC
	parentQuorumCert, err := s.reloadJustifyQC(parentID)
	if err != nil {
		logrus.Error("smr::ProcessProposal reloadJustifyQC error", "err", err)
		return err
	}
	parentQuorumCertBytes, err := json.Marshal(parentQuorumCert)
	if err != nil {
		return err
	}
	proposal := &pb.ProposalMsg{
		ProposalView: viewNumber,
		ProposalId:   proposalID,
		Timestamp:    time.Now().UnixNano(),
		JustifyQC:    parentQuorumCertBytes,
	}
	propMsg, err := s.cryptoClient.SignProposalMsg(proposal)
	if err != nil {
		logrus.Error("smr::ProcessProposal SignProposalMsg error", "error", err)
		return err
	}
	netMsg := NewMessage(pb.Message_NEW_PROPOSAL, propMsg, WithBCName(s.bcName))
	// 全部预备之后，再调用该接口
	if netMsg == nil {
		logrus.Error("smr::ProcessProposal::NewMessage error")
		return ErrP2PInternalErr
	}

	//go s.p2p.SubP2P(MsgTypeToString(pb.MsgType_NEW_VIEW))
	go s.p2p.PubP2P(MsgTypeToString(pb.MsgType_NEW_PROPOSAL), proto.Marshal(netMsg))
	logrus.Debug("smr::ProcessProposal::proposal", "localAddress", s.address, "validatesIpInfo", validatesIpInfo,
		"ProposalView", proposal.ProposalView, "ProposalId", fmt.Sprintf("%x", proposal.ProposalId),
		"Timestamp", proposal.Timestamp, "JustifyQC", proposal.JustifyQC)

	s.localProposal.Store(fmt.Sprintf("%x", proposalID), proposal.Timestamp)
	// 若为单候选人情况，则此处需要特殊处理，矿工需要给自己提前签名
	if len(validatesIpInfo) == 1 {
		s.voteToSelf(viewNumber, proposalID, parentQuorumCert)
	}
	logrus.Debug("smr:ProcessProposal::new proposal has been made", "address", s.address, "proposalID", fmt.Sprintf("%x", proposalID), "target", validatesIpInfo)
	return nil
}

func (s *Smr) voteToSelf(viewNumber int64, proposalID []byte, parent storage.QuorumCertInterface) {
	selfVote := &storage.VoteInfo{
		ProposalId:   proposalID,
		ProposalView: viewNumber,
		ParentId:     parent.GetProposalId(),
	}
	selfLedgerInfo := &storage.LedgerCommitInfo{
		VoteInfoHash: proposalID,
	}
	selfQC := storage.NewQuorumCert(selfVote, selfLedgerInfo, nil)
	selfSign, err := s.cryptoClient.SignVoteMsg(proposalID)
	if err != nil {
		logrus.Error("smr::voteProposal::voteToSelf error", "err", err)
		return
	}
	s.qcVoteMsgs.LoadOrStore(fmt.Sprintf("%x", proposalID), []*pb.QuorumCertSign{selfSign})
	selfNode := &storage.ProposalNode{
		In: selfQC,
	}
	if err := s.qcTree.UpdateQcStatus(selfNode); err != nil {
		logrus.Error("smr::voteProposal::updateQcStatus error", "err", err)
		return
	}
	// 更新本地smr状态机
	s.pacemaker.AdvanceView(selfQC)
	s.qcTree.UpdateHighQC(proposalID)
	logrus.Debug("smr:voteProposal::done local voting", "address", s.address, "proposalID", fmt.Sprintf("%x", proposalID))
}

// reloadJustifyQC 与LibraBFT不同，返回一个指定的parentQC
func (s *Smr) reloadJustifyQC(parentID []byte) (storage.QuorumCertInterface, error) {
	// 第一次proposal，highQC==rootQC==genesisQC
	if bytes.Equal(s.qcTree.GetGenesisQC().In.GetProposalId(), parentID) {
		highQC := s.getHighQC()
		return highQC, nil
	}
	// 若当前找不到，可能是qcTree已经更新了，废弃
	qc := s.qcTree.DFSQueryNode(parentID)
	if qc == nil {
		return nil, ErrEmptyTarget
	}
	v := &storage.VoteInfo{
		ProposalView: qc.In.GetProposalView(),
		ProposalId:   qc.In.GetProposalId(),
	}
	// 查看qcTree是否包含当前可以commit的Id
	var commitId []byte
	if s.qcTree.GetCommitQC() != nil {
		commitId = s.qcTree.GetCommitQC().In.GetProposalId()
	}

	// 根据qcTree生成一个parentQC
	// 上一个view的votes
	value, ok := s.qcVoteMsgs.Load(fmt.Sprintf("%x", v.ProposalId))
	if !ok {
		return nil, ErrJustifyVotesEmpty
	}
	signs, _ := value.([]*pb.QuorumCertSignature)
	parentQuorumCert := storage.NewQuorumCert(v, &storage.LedgerCommitInfo{
		CommitStateId: commitId,
	}, signs)
	return parentQuorumCert, nil
}

// handleReceivedProposal 该阶段在收到一个ProposalMsg后触发，与LibraBFT的process_proposal阶段类似
// 该阶段分两个角色，一个是认为自己是currentRound的Leader，一个是Replica
// 0. 查看ProposalMsg消息的合法性
// 1. 检查新的view是否符合账本状态要求
// 2. 比较本地pacemaker是否需要AdvanceRound
// 3. 检查qcTree是否需要更新CommitQC
// 4. 查看收到的view是否符合要求
// 5. 向本地PendingTree插入该QC，即更新QC
// 6. 发送一个vote消息给下一个Leader
// 注意：该过程删除了当前round的leader是否符合计算，将该步骤后置到上层共识CheckMinerMatch，原因：需要支持上层基于时间调度而不是基于round调度，减小耦合
func (s *Smr) handleReceivedProposal(msg *proto.Message) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	newProposalMsg := &pb.ProposalMsg{}
	if err := Unmarshal(msg, newProposalMsg); err != nil {
		logrus.Error("smr::handleReceivedProposal Unmarshal msg error", "logid", msg.GetHeader().GetLogid(), "error", err)
		return
	}

	_, ok := s.localProposal.LoadOrStore(fmt.Sprintf("%x", newProposalMsg.GetProposalId()), newProposalMsg.Timestamp)
	if ok && newProposalMsg.GetSign().Address != s.address {
		return
	}

	logrus.Debug("smr::handleReceivedProposal::received a proposal", "logid", msg.GetHeader().GetLogid(),
		"newView", newProposalMsg.GetProposalView(), "newProposalId", fmt.Sprintf("%x", newProposalMsg.GetProposalId()))
	parentQCBytes := newProposalMsg.GetJustifyQC()
	parentQC := &storage.QuorumCert{}
	if err := json.Unmarshal(parentQCBytes, parentQC); err != nil {
		logrus.Error("smr::handleReceivedProposal Unmarshal parentQC error", "error", err)
		return
	}

	newVote := &storage.VoteInfo{
		ProposalId:   newProposalMsg.GetProposalId(),
		ProposalView: newProposalMsg.GetProposalView(),
		ParentId:     parentQC.GetProposalId(),
		ParentView:   parentQC.GetProposalView(),
	}
	isFirstJustify := bytes.Equal(s.qcTree.GetGenesisQC().In.GetProposalId(), parentQC.GetProposalId())
	// 0.若为初始状态，则无需检查justify，否则需要检查qc有效性
	if !isFirstJustify {
		proposalQC := storage.NewQuorumCert(newVote, nil, []*pb.QuorumCertSignature{newProposalMsg.GetSign()})
		if err := s.saftyrules.CheckProposal(proposalQC, parentQC, s.election.GetValidators(parentQC.GetProposalView())); err != nil {
			logrus.Debug("smr::handleReceivedProposal::CheckProposal error", "error", err,
				"parentView", parentQC.GetProposalView(), "parentId", fmt.Sprintf("%x", parentQC.GetProposalId()))
			return
		}
	}
	// 1.检查账本状态和收到新round是否符合要求
	if s.ledgerState+StrictInternal < newVote.ProposalView {
		logrus.Error("smr::handleReceivedProposal::local IBlockledger hasn't been updated.", "LedgerState", s.ledgerState, "ProposalView", newVote.ProposalView)
		return
	}
	// 2.本地pacemaker试图更新currentView, 并返回一个是否需要将新消息通知该轮Leader, 是该轮不是下轮！主要解决P2PIP端口不能通知Loop的问题
	sendMsg, _ := s.pacemaker.AdvanceView(parentQC)
	logrus.Debug("smr::handleReceivedProposal::pacemaker update", "view", s.pacemaker.GetCurrentView())
	// 通知current Leader
	if sendMsg {
		netMsg := NewMessage(pb.Message_NEW_PROPOSAL, newProposalMsg, WithBCName(s.bcName))
		leader := newProposalMsg.GetSign().GetAddress()
		// 此处如果失败，仍会执行下层逻辑，因为是多个节点通知该轮Leader，因此若发不出去仍可继续运行
		if leader != "" && netMsg != nil && leader != s.address {
			go s.p2p.PubP2P(MsgTypeToString(pb.MsgType_NEW_PROPOSAL), proto.Marshal(netMsg))
			logrus.Debug("smr::handleReceivedProposal::proposal", "localAddress", s.address, "leader", leader,
				"ProposalView", newProposalMsg.ProposalView, "ProposalId", fmt.Sprintf("%x", newProposalMsg.ProposalId),
				"Timestamp", newProposalMsg.Timestamp, "JustifyQC", newProposalMsg.JustifyQC)
		}
	}

	// 3.本地safetyrules更新, 如有可以commit的QC，执行commit操作并更新本地rootQC
	if parentQC.LedgerCommitInfo != nil && parentQC.LedgerCommitInfo.CommitStateId != nil &&
		s.saftyrules.UpdatePreferredRound(parentQC.GetProposalView()) {
		s.qcTree.UpdateCommit(parentQC.GetProposalId())
	}
	// 4.查看收到的view是否符合要求, 此处接受孤儿节点
	if !s.saftyrules.CheckPacemaker(newProposalMsg.GetProposalView(), s.pacemaker.GetCurrentView()) {
		logrus.Error("smr::handleReceivedProposal::error", "error", ErrTooLowNewProposal, "local want", s.pacemaker.GetCurrentView(),
			"proposal have", newProposalMsg.GetProposalView())
		return
	}

	// 注意：删除此处的验证收到的proposal是否符合local计算，在本账本状态中后置到上层共识CheckMinerMatch
	// 根据本地saftyrules返回是否 需要发送voteMsg给下一个Leader
	if !s.saftyrules.VoteProposal(newProposalMsg.GetProposalId(), newProposalMsg.GetProposalView(), parentQC) {
		logrus.Error("smr::handleReceivedProposal::VoteProposal fail", "view", newProposalMsg.GetProposalView(), "proposalId", newProposalMsg.GetProposalId())
		return
	}

	// 这个newVoteId表示的是本地最新一次vote的id，生成voteInfo的hash，标识vote消息
	newLedgerInfo := &storage.LedgerCommitInfo{
		VoteInfoHash: newProposalMsg.GetProposalId(),
	}
	newNode := &storage.ProposalNode{
		In: storage.NewQuorumCert(newVote, newLedgerInfo, nil),
	}
	// 5.与proposal.ParentId相比，更新本地qcTree，insert新节点, 包括更新CommitQC等等
	if err := s.qcTree.UpdateQcStatus(newNode); err != nil {
		logrus.Error("smr::handleReceivedProposal::updateQcStatus error", "err", err)
		return
	}
	logrus.Debug("smr::handleReceivedProposal::pacemaker changed", "round", s.pacemaker.GetCurrentView())
	// 6.发送一个vote消息给下一个Leader
	nextLeader := s.election.GetLeader(s.pacemaker.GetCurrentView() + 1)
	if nextLeader == "" {
		logrus.Warn("smr::handleReceivedProposal::empty next leader", "next round", s.pacemaker.GetCurrentView()+1)
		return
	}
	s.voteProposal(newProposalMsg.GetProposalId(), newVote, newLedgerInfo, nextLeader)
}

// voteProposal 当Replica收到一个Proposal并对该Proposal检查之后，该节点会针对该QC投票
// 节点的vote包含一个本次vote的对象的基本信息，和本地上次vote对象的基本信息，和本地账本的基本信息，和一个签名
// 只要vote过，就在本地map中更新值
func (s *Smr) voteProposal(msg []byte, vote *storage.VoteInfo, ledger *storage.LedgerCommitInfo, voteTo string) {
	// 若为自己直接先返回
	if voteTo == s.address {
		return
	}
	nextSign, err := s.cryptoClient.SignVoteMsg(msg)
	if err != nil {
		logrus.Error("smr::voteProposal::SignVoteMsg error", "err", err)
		return
	}
	voteBytes, err := json.Marshal(vote)
	if err != nil {
		logrus.Error("smr::voteProposal::Marshal vote error", "err", err)
		return
	}
	ledgerBytes, err := json.Marshal(ledger)
	if err != nil {
		logrus.Error("smr::voteProposal::Marshal commit error", "err", err)
		return
	}
	voteMsg := &pb.VoteMsg{
		VoteInfo:         voteBytes,
		LedgerCommitInfo: ledgerBytes,
		Signature:        []*pb.QuorumCertSignature{nextSign},
	}
	netMsg := NewMessage(pb.Message_VOTE, voteMsg, WithBCName(s.bcName))
	// 全部预备之后，再调用该接口
	if netMsg == nil {
		logrus.Error("smr::ProcessProposal::NewMessage error")
		return
	}
	go s.p2p.PubP2P(MsgTypeToString(pb.MsgType_VOTE), proto.Marshal(netMsg))
	logrus.Debug("smr::voteProposal::vote", "vote to next leader", voteTo, "vote view number", vote.ProposalView)
}

// handleReceivedVoteMsg 当前Leader在发送一个proposal消息之后，由下一Leader等待周围replica的投票，收集vote消息
// 当收到2f+1个vote消息之后，本地pacemaker调用AdvanceView，并更新highQC
// 该方法针对Leader而言
func (s *Smr) handleReceivedVoteMsg(msg *pb.Message) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	newVoteMsg := &pb.VoteMsg{}
	if err := Unmarshal(msg, newVoteMsg); err != nil {
		logrus.Error("smr::handleReceivedVoteMsg Unmarshal msg error", "logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}
	voteQC, err := s.voteMsgToQC(newVoteMsg)
	if err != nil {
		logrus.Error("smr::handleReceivedVoteMsg VoteMsgToQC error", "error", err)
		return err
	}
	// 检查logid、voteInfoHash是否正确
	if err := s.saftyrules.CheckVote(voteQC, msg.GetHeader().GetLogid(), s.election.GetValidators(voteQC.GetProposalView())); err != nil {
		logrus.Error("smr::handleReceivedVoteMsg CheckVote error", "error", err, "msg", fmt.Sprintf("%x", voteQC.GetProposalId()))
		return err
	}
	logrus.Debug("smr::handleReceivedVoteMsg::receive vote", "voteId", fmt.Sprintf("%x", voteQC.GetProposalId()), "voteView", voteQC.GetProposalView(), "from", voteQC.GetSignsInfo()[0].Address)

	// 若vote先于proposal到达，则直接丢弃票数
	if _, ok := s.localProposal.Load(fmt.Sprintf("%x", voteQC.GetProposalId())); !ok {
		logrus.Debug("smr::handleReceivedVoteMsg::haven't received the related proposal msg, drop it.")
		return ErrEmptyTarget
	}
	if node := s.qcTree.DFSQueryNode(voteQC.GetProposalId()); node == nil {
		logrus.Debug("smr::handleReceivedVoteMsg::haven't finish proposal process, drop it.")
		return ErrEmptyTarget
	}

	// 存入本地voteInfo内存，查看签名数量是否超过2f+1
	var VoteLen int
	// 注意隐式，若!ok则证明签名数量为1，此时不可能超过2f+1
	v, ok := s.qcVoteMsgs.LoadOrStore(fmt.Sprintf("%x", voteQC.GetProposalId()), voteQC.GetSignsInfo())
	// 若ok=false，则仅store一个vote签名
	VoteLen = 1
	if ok {
		signs, _ := v.([]*pb.QuorumCertSignature)
		stored := false
		for _, sign := range signs {
			// 自己给自己投票将自动忽略
			if sign.Address == voteQC.GetSignsInfo()[0].Address || voteQC.GetSignsInfo()[0].Address == s.address {
				stored = true
			}
		}
		if !stored {
			signs = append(signs, voteQC.GetSignsInfo()[0])
			s.qcVoteMsgs.Store(fmt.Sprintf("%x", voteQC.GetProposalId()), signs)
		}
		VoteLen = len(signs)
	}
	// 查看签名数量是否达到2f+1, 需要获取justify对应的validators
	if !s.saftyrules.CalVotesThreshold(VoteLen, len(s.election.GetValidators(voteQC.GetProposalView()))) {
		return nil
	}

	// 更新本地pacemaker AdvanceRound
	s.pacemaker.AdvanceView(voteQC)
	logrus.Debug("smr::handleReceivedVoteMsg::FULL VOTES!", "pacemaker view", s.pacemaker.GetCurrentView())
	// 更新HighQC
	s.qcTree.UpdateHighQC(voteQC.GetProposalId())
	return nil
}

// voteMsgToQC 提供一个从VoteMsg转化为quorumCert的方法，注意，两者struct其实相仿
func (s *Smr) voteMsgToQC(msg *pb.VoteMsg) (storage.QuorumCertInterface, error) {
	voteInfo := &storage.VoteInfo{}
	if err := json.Unmarshal(msg.VoteInfo, voteInfo); err != nil {
		return nil, err
	}
	ledgerCommitInfo := &storage.LedgerCommitInfo{}
	if err := json.Unmarshal(msg.LedgerCommitInfo, ledgerCommitInfo); err != nil {
		return nil, err
	}
	return storage.NewQuorumCert(voteInfo, ledgerCommitInfo, msg.GetSignature()), nil
}

func (s *Smr) blockToProposalNode(block IBlock) *storage.ProposalNode {
	targetId := block.GetBlockid()
	if node := s.qcTree.DFSQueryNode(targetId); node != nil {
		return node
	}
	v := &storage.VoteInfo{
		ProposalId:   block.GetBlockid(),
		ProposalView: block.GetHeight(),
		ParentId:     block.GetPreHash(),
		ParentView:   block.GetHeight() - 1,
	}
	return &storage.ProposalNode{In: storage.NewQuorumCert(v, nil, nil)}
}

func (s *Smr) getHighQC() storage.QuorumCertInterface {
	return s.qcTree.GetHighQC().In
}

// getCompleteHighQC 本地qcTree不带签名，因此smr需要重新组装完整的QC
func (s *Smr) getCompleteHighQC() storage.QuorumCertInterface {
	raw := s.getHighQC()
	vote := &storage.VoteInfo{
		ProposalId:   raw.GetProposalId(),
		ProposalView: raw.GetProposalView(),
		ParentId:     raw.GetParentProposalId(),
		ParentView:   raw.GetProposalView(),
	}
	signInfo, ok := s.qcVoteMsgs.Load(fmt.Sprintf("%x", raw.GetProposalId()))
	if !ok {
		return storage.NewQuorumCert(vote, nil, nil)
	}
	signs, _ := signInfo.([]*pb.QuorumCertSignature)
	return storage.NewQuorumCert(vote, nil, signs)
}

func (s *Smr) validNewHighQC(inProposalId []byte, validators []string) bool {
	signInfo, ok := s.qcVoteMsgs.Load(fmt.Sprintf("%x", inProposalId))
	if !ok {
		return false
	}
	signs, ok := signInfo.([]*pb.QuorumCertSignature)
	if !ok {
		return false
	}
	if len(validators) == 1 {
		return len(signs) == len(validators)
	}
	return s.saftyrules.CalVotesThreshold(len(signs), len(validators))
}

func (s *Smr) enforceUpdateHighQC(inProposalId []byte) (bool, error) {
	if bytes.Equal(s.getHighQC().GetProposalId(), inProposalId) {
		return false, nil
	}
	return true, s.qcTree.EnforceUpdateHighQC(inProposalId)
}

func (s *Smr) removeLocalValidator(in []string) []string {
	var out []string
	for _, addr := range in {
		if addr != s.address {
			out = append(out, addr)
		}
	}
	return out
}

// func createNewBCtx() *xctx.BaseCtx {
// 	log, _ := logs.NewLogger("", "smr")
// 	return &xctx.BaseCtx{
// 		XLog:  log,
// 		Timer: timer.NewXTimer(),
// 	}
// }

// appendSigns 将p中不重复的签名append进q中
func appendSigns(q []*pb.QuorumCertSignature, p []*pb.QuorumCertSignature) []*pb.QuorumCertSignature {
	signSet := make(map[string]bool)
	for _, sign := range q {
		if _, ok := signSet[sign.Address]; !ok {
			signSet[sign.Address] = true
		}
	}
	for _, sign := range p {
		if _, ok := signSet[sign.Address]; !ok {
			q = append(q, sign)
		}
	}
	return q
}
