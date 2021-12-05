// Copyright Xuperchain Authors
// link: https://github.com/xuperchain/xupercore

package chained_hotstuff

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/core/keypair"
)

var (
	EmptyVoteSignErr   = errors.New("No signature in vote.")
	InvalidVoteAddr    = errors.New("Vote address is not a validator in the target validators.")
	InvalidVoteSign    = errors.New("Vote sign is invalid compared with its publicKey")
	TooLowVoteView     = errors.New("Vote received is lower than local lastVoteRound.")
	TooLowVParentView  = errors.New("Vote's parent received is lower than local preferredRound.")
	TooLowProposalView = errors.New("Proposal received is lower than local lastVoteRound.")
	EmptyParentQC      = errors.New("Parent qc is empty.")
	NoEnoughVotes      = errors.New("Parent qc doesn't have enough votes.")
	EmptyParentNode    = errors.New("Parent's node is empty.")
	EmptyValidators    = errors.New("Justify validators are empty.")
)

type iSaftyRules interface {
	UpdatePreferredRound(round int64) bool
	VoteProposal(proposalId []byte, proposalRound int64, parentQc IQuorumCert) bool
	CheckVote(qc IQuorumCert, validators []string) error
	CalVotesThreshold(input, sum int) bool
	CheckProposal(proposal, parent IQuorumCert, justifyValidators []string) error
	CheckPacemaker(pending, local int64) bool
}

type DefaultSaftyRules struct {
	// lastVoteRound 存储着本地最近一次投票的轮数
	lastVoteRound int64
	// preferredRound 存储着本地PendingTree
	// 即有[两个子孙节点的节点]
	// 若本地有相同高度的节点，则自然排序后选出preferredRound
	preferredRound int64
	Pubkey         keypair.PubKey
	QcTree         *QCPendingTree
}

func (s *DefaultSaftyRules) UpdatePreferredRound(round int64) bool {
	if round-1 > s.preferredRound {
		s.preferredRound = round - 1
	}
	// TODO: 检查LedgerInfo是否一致
	return true
}

// VoteProposal 返回是否需要发送voteMsg给下一个Leader
// DefaultSaftyRules 并没有严格比对proposalRound和parentRound的相邻自增关系
// 但需要注意的是，在上层bcs的实现中，由于共识操纵了账本回滚。因此实际上safetyrules需要proposalRound和parentRound严格相邻的
// 因此由于账本的可回滚性，因此lastVoteRound和preferredRound比对时，仅需比对新来的数据是否小于local数据-3即可
// 此处-3代表数据已经落盘
func (s *DefaultSaftyRules) VoteProposal(proposalId []byte, proposalRound int64, parentQc IQuorumCert) bool {
	if proposalRound < s.lastVoteRound-3 {
		return false
	}
	if parentQc.GetProposalView() < s.preferredRound-3 {
		return false
	}
	s.increaseLastVoteRound(proposalRound)
	return true
}

// CheckVote 检查logid、voteInfoHash是否正确
func (s *DefaultSaftyRules) CheckVote(qc IQuorumCert, validators []string) error {
	// 检查签名, vote目前为单个签名，因此只需要验证第一个即可，验证的内容为签名信息是否在合法的validators里面
	signs := qc.GetSignsInfo()
	if len(signs) == 0 {
		return EmptyVoteSignErr
	}
	// 是否是来自有效的候选人
	if !isInSlice(signs[0].GetAddress(), validators) {
		logrus.Errorf("DefaultSaftyRules::CheckVote Validators(%v) from(%s) error", validators, signs[0].GetAddress())
		return InvalidVoteAddr
	}
	// 签名和公钥是否匹配
	if ok := s.Pubkey.VerifySignature(qc.GetProposalId(), signs[0].GetSign()); !ok {
		return InvalidVoteSign
	}
	// 检查voteinfo信息, proposalView小于lastVoteRound，parentView不小于preferredRound
	if qc.GetProposalView() < s.lastVoteRound-3 {
		return TooLowVoteView
	}
	if qc.GetParentView() < s.preferredRound-3 {
		return TooLowVParentView
	}
	// TODO: 检查commit消息
	return nil
}

func (s *DefaultSaftyRules) increaseLastVoteRound(round int64) {
	if round > s.lastVoteRound {
		s.lastVoteRound = round
	}
}

func (s *DefaultSaftyRules) CalVotesThreshold(input, sum int) bool {
	// 计算最大恶意节点数, input+1表示去除自己的签名
	f := (sum - 1) / 3
	if f < 0 {
		return false
	}
	if f == 0 {
		return input+1 >= sum
	}
	return input+1 >= sum-f
}

// CheckProposalMsg 原IsQuorumCertValidate 判断justify，即需check的block的parentQC是否合法
// 需要注意的是，在上层bcs的实现中，由于共识操纵了账本回滚。因此实际上safetyrules需要proposalRound和parentRound严格相邻的
// 因此在此proposal和parent的QC稍微宽松检查
func (s *DefaultSaftyRules) CheckProposal(proposal, parent IQuorumCert, justifyValidators []string) error {
	if proposal.GetProposalView() < s.lastVoteRound-3 {
		return TooLowProposalView
	}
	if justifyValidators == nil {
		return EmptyValidators
	}
	// step2: verify justify's votes

	// verify justify sign number
	if parent.GetProposalId() == nil {
		return EmptyParentQC
	}

	// 新qc至少要在本地qcTree挂上, 那么justify的节点需要在本地
	// 或者新qc目前为孤儿节点，有可能未来切换成HighQC，此时仅需要proposal在[root+1, root+6]
	// 是+6不是+3的原因是考虑到重起的时候的情况，重起时，root为tipId-3，而外界状态最多到tipId+3，此处简化处理
	if parentNode := s.QcTree.DFSQueryNode(parent.GetProposalId()); parentNode == nil {
		if proposal.GetProposalView() <= s.QcTree.Root.In.GetParentView() || proposal.GetProposalView() > s.QcTree.Root.In.GetProposalView()+6 {
			return EmptyParentNode
		}
	}

	// 检查justify的所有vote签名
	justifySigns := parent.GetSignsInfo()
	validCnt := 0
	for _, v := range justifySigns {
		if !isInSlice(v.GetAddress(), justifyValidators) {
			continue
		}
		// 签名和公钥是否匹配
		if ok := s.Pubkey.VerifySignature(parent.GetProposalId(), v.GetSign()); !ok {
			return InvalidVoteSign
		}
		validCnt++
	}
	if !s.CalVotesThreshold(validCnt, len(justifyValidators)) {
		return NoEnoughVotes
	}
	return nil
}

// CheckPacemaker
// 注意： 由于本smr支持不同节点产生同一round， 因此下述round比较和leader比较与原文(验证Proposal的Round是否和pacemaker的Round相等)并不同。
// 仅需proposal round不超过范围即可
func (s *DefaultSaftyRules) CheckPacemaker(pending int64, local int64) bool {
	if pending <= local-3 {
		return false
	}
	return true
}

func isInSlice(target string, s []string) bool {
	for _, v := range s {
		if target == v {
			return true
		}
	}
	return false
}
