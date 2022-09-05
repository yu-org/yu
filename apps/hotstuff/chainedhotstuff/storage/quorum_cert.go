package storage

import (
	"errors"

    pb "github.com/yu-org/yu/apps/hotstuff/chainedhotstuff/proto"
)

var _ QuorumCertInterface = (*QuorumCert)(nil)

var (
	ErrNoValidQC       = errors.New("target qc is empty")
	ErrNoValidParentId = errors.New("parentId is empty")
)

// 本文件定义了chained-bft下有关的数据结构和接口
// QuorumCertInterface 规定了pacemaker和saftyrules操作的qc接口
// QuorumCert 为一个QuorumCertInterface的实现 TODO: smr彻底接口化是否可能?
// QCPendingTree 规定了smr内存存储的组织形式，其为一个QC树状结构

// IQuorumCert 接口
type IQuorumCert interface {
	GetProposalView() int64
	GetProposalId() []byte
	GetParentProposalId() []byte
	GetParentView() int64
	GetSignsInfo() []*pb.QuorumCertSignature
}

// VoteInfo 包含了本次和上次的vote对象
type VoteInfo struct {
	// 本次vote的对象
	ProposalId   []byte
	ProposalView int64
	// 本地上次vote的对象
	ParentId   []byte
	ParentView int64
}

// ledgerCommitInfo 表示的是本地账本和QC存储的状态，包含一个commitStateId和一个voteInfoHash
// commitStateId 表示本地账本状态，TODO: = 本地账本merkel root
// voteInfoHash 表示本地vote的vote_info的哈希，即本地QC的最新状态
type LedgerCommitInfo struct {
	CommitStateId []byte
	VoteInfoHash  []byte
}

func NewQuorumCert(v *VoteInfo, l *LedgerCommitInfo, s []*pb.QuorumCertSignature) QuorumCertInterface {
	qc := QuorumCert{
		VoteInfo:         v,
		LedgerCommitInfo: l,
		SignInfos:        s,
	}
	return &qc
}
