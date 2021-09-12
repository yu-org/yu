// Copyright Xuperchain Authors
// link: https://github.com/xuperchain/xupercore

package chained_hotstuff

import (
	"bytes"
	"container/list"
	"errors"
	chainedBftPb "github.com/xuperchain/xupercore/kernel/consensus/base/driver/chained-bft/pb"
	"github.com/xuperchain/xupercore/lib/utils"

	. "github.com/Lawliet-Chan/yu/yerror"
	"github.com/sirupsen/logrus"
)

var _ IQuorumCert = (*QuorumCert)(nil)

// 本文件定义了chained-bft下有关的数据结构和接口
// IQuorumCert 规定了pacemaker和saftyrules操作的qc接口
// QuorumCert 为一个IQuorumCert的实现
// QCPendingTree 规定了smr内存存储的组织形式，其为一个QC树状结构

// IQuorumCert 接口
type IQuorumCert interface {
	GetProposalView() int64
	GetProposalId() []byte
	GetParentProposalId() []byte
	GetParentView() int64
	GetSignsInfo() []*chainedBftPb.QuorumCertSign
}

// quorumCert 是HotStuff的基础结构，它表示了一个节点本地状态以及其余节点对该状态的确认
type QuorumCert struct {
	// 本次qc的vote对象，该对象中嵌入了上次的QCid，因此删除原有的ProposalMsg部分
	VoteInfo *VoteInfo
	// 当前本地账本的状态
	LedgerCommitInfo *LedgerCommitInfo
	// SignInfos is the signs of the leader gathered from replicas of a specifically certType.
	SignInfos []*chainedBftPb.QuorumCertSign
}

func (qc *QuorumCert) GetProposalView() int64 {
	return qc.VoteInfo.ProposalView
}

func (qc *QuorumCert) GetProposalId() []byte {
	return qc.VoteInfo.ProposalId
}

func (qc *QuorumCert) GetParentProposalId() []byte {
	return qc.VoteInfo.ParentId
}

func (qc *QuorumCert) GetParentView() int64 {
	return qc.VoteInfo.ParentView
}

func (qc *QuorumCert) GetSignsInfo() []*chainedBftPb.QuorumCertSign {
	return qc.SignInfos
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

// PendingTree 是一个内存内的QC状态存储树，仅存放目前未commit(即可能触发账本回滚)的区块信息
// 当PendingTree中的某个节点有[严格连续的]三代子孙后，将出发针对该节点的账本Commit操作
// 本数据结构替代原有Chained-BFT的三层QC存储，即proposalQC,generateQC和lockedQC
type QCPendingTree struct {
	Genesis   *ProposalNode // Tree中第一个Node
	Root      *ProposalNode
	HighQC    *ProposalNode // Tree中最高的QC指针
	GenericQC *ProposalNode
	LockedQC  *ProposalNode
	CommitQC  *ProposalNode

	OrphanList *list.List // []*ProposalNode孤儿数组
	OrphanMap  map[string]bool
}

type ProposalNode struct {
	In IQuorumCert
	// Parent IQuorumCert
	Sons []*ProposalNode
	// Parent *ProposalNode
}

func (t *QCPendingTree) GetRootQC() *ProposalNode {
	return t.Root
}

func (t *QCPendingTree) GetHighQC() *ProposalNode {
	return t.HighQC
}

func (t *QCPendingTree) GetGenericQC() *ProposalNode {
	return t.GenericQC
}

func (t *QCPendingTree) GetCommitQC() *ProposalNode {
	return t.CommitQC
}

func (t *QCPendingTree) GetLockedQC() *ProposalNode {
	return t.LockedQC
}

// 更新本地qcTree, insert新节点, 将新节点parentQC和本地HighQC对比，如有必要进行更新
func (t *QCPendingTree) updateQcStatus(node *ProposalNode) error {
	if t.DFSQueryNode(node.In.GetProposalId()) != nil {
		logrus.Debug("QCPendingTree::updateQcStatus::has been inserted", "search", utils.F(node.In.GetProposalId()))
		return nil
	}
	if err := t.insert(node); err != nil {
		logrus.Error("QCPendingTree::updateQcStatus insert err", "err", err)
		return err
	}
	t.updateHighQC(node.In.GetParentProposalId())
	logrus.Debug("QCPendingTree::updateQcStatus", "insert new", utils.F(node.In.GetProposalId()), "height", node.In.GetProposalView(), "highQC", utils.F(t.GetHighQC().In.GetProposalId()))
	return nil
}

// updateHighQC 对比QC树，将本地HighQC和输入id比较，高度更高的更新为HighQC，此时连同GenericQC、LockedQC、CommitQC一起修改
func (t *QCPendingTree) updateHighQC(inProposalId []byte) {
	node := t.DFSQueryNode(inProposalId)
	if node == nil {
		logrus.Debug("QCPendingTree::updateHighQC::DFSQueryNode nil!", "id", utils.F(inProposalId))
		return
	}
	// 若新验证过的node和原HighQC高度相同，使用新验证的node
	if node.In.GetProposalView() < t.GetHighQC().In.GetProposalView() {
		return
	}
	// 更改HighQC以及一系列的GenericQC、LockedQC和CommitQC
	t.HighQC = node
	logrus.Debug("QCPendingTree::updateHighQC", "HighQC height", node.In.GetProposalView(), "HighQC", utils.F(node.In.GetProposalId()))
	parent := t.DFSQueryNode(node.In.GetParentProposalId())
	if parent == nil {
		return
	}
	t.GenericQC = parent
	logrus.Debug("QCPendingTree::updateHighQC", "GenericQC height", t.GenericQC.In.GetProposalView(), "GenericQC", utils.F(t.GenericQC.In.GetProposalId()))
	// 找grand节点，标为LockedQC
	parentParent := t.DFSQueryNode(parent.In.GetParentProposalId())
	if parentParent == nil {
		return
	}
	t.LockedQC = parentParent
	logrus.Debug("QCPendingTree::updateHighQC", "LockedQC height", t.LockedQC.In.GetProposalView(), "LockedQC", utils.F(t.LockedQC.In.GetProposalId()))
	// 找grandgrand节点，标为CommitQC
	parentParentParent := t.DFSQueryNode(parentParent.In.GetParentProposalId())
	if parentParentParent == nil {
		return
	}
	t.CommitQC = parentParentParent
	logrus.Debug("QCPendingTree::updateHighQC", "CommitQC height", t.CommitQC.In.GetProposalView(), "CommitQC", utils.F(t.CommitQC.In.GetProposalId()))
}

// enforceUpdateHighQC 强制更改HighQC指针，用于错误时回滚，注意: 本实现没有timeoutQC因此需要此方法
func (t *QCPendingTree) enforceUpdateHighQC(inProposalId []byte) error {
	node := t.DFSQueryNode(inProposalId)
	if node == nil {
		logrus.Debug("QCPendingTree::enforceUpdateHighQC::DFSQueryNode nil")
		return NoValidQC
	}
	// 更改HighQC以及一系列的GenericQC、LockedQC和CommitQC
	t.HighQC = node
	t.GenericQC = nil
	t.LockedQC = nil
	t.CommitQC = nil
	logrus.Debug("QCPendingTree::enforceUpdateHighQC", "HighQC height", t.HighQC.In.GetProposalView(), "HighQC", utils.F(t.HighQC.In.GetProposalId()))
	parent := t.DFSQueryNode(node.In.GetParentProposalId())
	if parent == nil {
		return nil
	}
	t.GenericQC = parent
	logrus.Debug("QCPendingTree::enforceUpdateHighQC", "GenericQC height", t.GenericQC.In.GetProposalView(), "GenericQC", utils.F(t.GenericQC.In.GetProposalId()))
	// 找grand节点，标为LockedQC
	parentParent := t.DFSQueryNode(parent.In.GetParentProposalId())
	if parentParent == nil {
		return nil
	}
	t.LockedQC = parentParent
	logrus.Debug("QCPendingTree::enforceUpdateHighQC", "LockedQC height", t.LockedQC.In.GetProposalView(), "LockedQC", utils.F(t.LockedQC.In.GetProposalId()))
	// 找grandgrand节点，标为Commit	QC
	parentParentParent := t.DFSQueryNode(parentParent.In.GetParentProposalId())
	if parentParentParent == nil {
		return nil
	}
	t.CommitQC = parentParentParent
	logrus.Debug("QCPendingTree::enforceUpdateHighQC", "CommitQC height", t.CommitQC.In.GetProposalView(), "CommitQC", utils.F(t.CommitQC.In.GetProposalId()))
	return nil
}

// insert 向本地QC树Insert一个ProposalNode，如有必要，连同HighQC、GenericQC、LockedQC、CommitQC一起修改
func (t *QCPendingTree) insert(node *ProposalNode) error {
	if node.In.GetParentProposalId() == nil {
		return NoValidParentId
	}
	parent := t.DFSQueryNode(node.In.GetParentProposalId())
	if parent != nil {
		parent.Sons = append(parent.Sons, node)
		t.adoptOrphans(node)
		return nil
	}
	// 作为孤儿节点加入
	t.insertOrphan(node)
	return nil
}

// insertOrphan为向孤儿数组插入孤儿节点的逻辑
// 若该node的父节点不存在在slice中，则查看该node的是否为slice中节点的父节点，若是则代替该节点反转挂上，若否继续查看
// 若该node的父节点存在在sli中，则直接挂在父节点下，如否则在sli中追加节点
// [A1,   B1,  C1,  D1 ...]
//  ｜    ||
//  A2  B2 B2'
//  |
//  A3
func (t *QCPendingTree) insertOrphan(node *ProposalNode) error {
	if _, ok := t.OrphanMap[utils.F(node.In.GetProposalId())]; ok {
		return nil // 重复退出
	}
	t.OrphanMap[utils.F(node.In.GetProposalId())] = true
	if t.OrphanList.Len() == 0 {
		t.OrphanList.PushBack(node)
		return nil
	}
	// 遍历整个Sli，查看是否能够挂上
	ptr := t.OrphanList.Front()
	for ptr != nil {
		curPtr := ptr
		n, ok := curPtr.Value.(*ProposalNode)
		if !ok {
			return errors.New("QCPendingTree::insertOrphan::element type invalid.")
		}
		ptr = ptr.Next()
		// 查看头节点是否已经时间失效了, 失效的时候所有依赖该高度长得树实际上都没有意义了，需要删除
		if n.In.GetProposalView() <= t.Root.In.GetProposalView() {
			t.OrphanList.Remove(curPtr)
			continue
		}
		// 查看头节点是否是node的儿子, 直接在头部插入
		if bytes.Equal(n.In.GetParentProposalId(), node.In.GetProposalId()) {
			node.Sons = append(node.Sons, n)
			t.OrphanList.Remove(curPtr)
			t.OrphanList.PushBack(node)
			return nil
		}
		// 否则遍历该树试图挂在子树上面
		parent := DFSQuery(n, node.In.GetParentProposalId())
		if parent != nil {
			parent.Sons = append(parent.Sons, node)
			return nil
		}
	}
	// 没有可以挂的地方，则直接append
	t.OrphanList.PushBack(node)
	return nil
}

// adoptOrphans 查看孤儿节点列表是否可以挂在该节点上
func (t *QCPendingTree) adoptOrphans(node *ProposalNode) error {
	if t.OrphanList.Len() == 0 {
		return nil
	}
	ptr := t.OrphanList.Front()
	for ptr != nil {
		curPtr := ptr
		n, ok := curPtr.Value.(*ProposalNode)
		if !ok {
			return errors.New("QCPendingTree::insertOrphan::element type invalid.")
		}
		ptr = ptr.Next()
		if bytes.Equal(n.In.GetParentProposalId(), node.In.GetProposalId()) {
			node.Sons = append(node.Sons, n)
			t.OrphanList.Remove(curPtr)
		}
	}
	return nil
}

/*
func (t *QCPendingTree) updateCommit(p IQuorumCert) {
	// t.Ledger.ConsensusCommit(p.GetProposalId())
	node := t.DFSQueryNode(p.GetProposalId())
	parent := node.Parent
	node.Parent = nil
	if parent != nil {
		parent.Sons = nil
	}
	t.Root = node
}
*/

// updateCommit 此方法向存储接口发送一个ProcessCommit，通知存储落盘，此时的block将不再被回滚
// 同时此方法将原先的root更改为commit node，因为commit node在本BFT中已确定不会回滚
func (t *QCPendingTree) updateCommit(id []byte) {
	node := t.DFSQueryNode(id)
	if node == nil {
		return
	}
	parent := t.DFSQueryNode(node.In.GetParentProposalId())
	if parent == nil {
		return
	}
	parentParent := t.DFSQueryNode(parent.In.GetParentProposalId())
	if parentParent == nil {
		return
	}
	parentParentParent := t.DFSQueryNode(parentParent.In.GetParentProposalId())
	if parentParentParent == nil {
		return
	}
	parentParentParentParent := t.DFSQueryNode(parentParentParent.In.GetParentProposalId())
	if parentParentParentParent == nil {
		return
	}
	parentParentParentParent.Sons = nil
	t.Root = parentParentParent
	// TODO: commitQC/lockedQC/genericQC/highQC是否有指向原root及以上的Node
}

// DFSQueryNode实现的比较简单，从root节点开始寻找，后续有更优方法可优化
func (t *QCPendingTree) DFSQueryNode(id []byte) *ProposalNode {
	return DFSQuery(t.Root, id)
}

func DFSQuery(node *ProposalNode, target []byte) *ProposalNode {
	if target == nil || node == nil {
		return nil
	}
	if bytes.Equal(node.In.GetProposalId(), target) {
		return node
	}
	if node.Sons == nil {
		return nil
	}
	for _, node := range node.Sons {
		if n := DFSQuery(node, target); n != nil {
			return n
		}
	}
	return nil
}
