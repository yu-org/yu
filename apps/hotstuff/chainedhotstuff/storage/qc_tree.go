package storage

import (
    "fmt"
	"bytes"
	"container/list"
	"errors"
	"sync"

	cctx "github.com/xuperchain/xupercore/kernel/consensus/context"
    "github.com/sirupsen/logrus"
)

type ProposalNode struct {
	In QuorumCertInterface
	// Parent QuorumCertInterface
	Sons []*ProposalNode
	// Parent *ProposalNode
}

type LedgerRely interface {
	GetConsensusConf() ([]byte, error)
	QueryBlockHeader(blkId []byte) (ledger.BlockHandle, error)
	QueryBlockHeaderByHeight(int64) (ledger.BlockHandle, error)
	GetTipBlock() ledger.BlockHandle
	GetTipXMSnapshotReader() (ledger.XMSnapshotReader, error)
	CreateSnapshot(blkId []byte) (ledger.XMReader, error)
	GetTipSnapshot() (ledger.XMReader, error)
	QueryTipBlockHeader() ledger.BlockHandle
}


func NewTreeNode(ledger LedgerRely, height int64) *ProposalNode {
	b, err := ledger.QueryBlockHeaderByHeight(height)
	if err != nil {
		return nil
	}
	pre, err := ledger.QueryBlockHeaderByHeight(height - 1)
	vote := VoteInfo{
		ProposalId:   b.GetBlockid(),
		ProposalView: b.GetHeight(),
	}
	ledgerInfo := LedgerCommitInfo{
		CommitStateId: b.GetBlockid(),
	}
	if err != nil {
		return &ProposalNode{
			In: NewQuorumCert(&vote, &ledgerInfo, nil),
		}
	}
	vote.ParentId = pre.GetBlockid()
	vote.ParentView = pre.GetHeight()
	return &ProposalNode{
		In: NewQuorumCert(&vote, &ledgerInfo, nil),
	}
}

// PendingTree 是一个内存内的QC状态存储树，仅存放目前未commit(即可能触发账本回滚)的区块信息
// 当PendingTree中的某个节点有[严格连续的]三代子孙后，将出发针对该节点的账本Commit操作
// 本数据结构替代原有Chained-BFT的三层QC存储，即proposalQC,generateQC和lockedQC
type QCPendingTree struct {
	genesis   *ProposalNode // Tree中第一个Node
	root      *ProposalNode
	highQC    *ProposalNode // Tree中最高的QC指针
	genericQC *ProposalNode
	lockedQC  *ProposalNode
	commitQC  *ProposalNode

	orphanList *list.List // []*ProposalNode孤儿数组
	orphanMap  map[string]bool

	mtx sync.RWMutex

	log logs.Logger
}

func MockTree(genesis *ProposalNode, root *ProposalNode, highQC *ProposalNode,
	genericQC *ProposalNode, lockedQC *ProposalNode, commitQC *ProposalNode,
	log logs.Logger) *QCPendingTree {
	return &QCPendingTree{
		genesis:    genesis,
		root:       root,
		highQC:     highQC,
		genericQC:  genericQC,
		lockedQC:   lockedQC,
		commitQC:   commitQC,
		log:        log,
		orphanList: list.New(),
		orphanMap:  make(map[string]bool),
	}
}

// initQCTree 创建了smr需要的QC树存储，该Tree存储了目前待commit的QC信息
func InitQCTree(startHeight int64, ledger cctx.LedgerRely) *QCPendingTree {
	// 初始状态应该是start高度的前一个区块为genesisQC，即tipBlock
	g, err := ledger.QueryBlockHeaderByHeight(startHeight - 1)
	if err != nil {
		logrus.Warn("InitQCTree QueryBlockHeaderByHeight failed", "error", err.Error())
		return nil
	}
	gQC := NewQuorumCert(
		&VoteInfo{
			ProposalId:   g.GetBlockid(),
			ProposalView: g.GetHeight(),
		},
		&LedgerCommitInfo{
			CommitStateId: g.GetBlockid(),
		},
		nil)
	gNode := &ProposalNode{
		In: gQC,
	}
	tip := ledger.GetTipBlock()
	// 当前为初始状态
	if tip.GetHeight() <= startHeight {
		return &QCPendingTree{
			genesis:    gNode,
			root:       gNode,
			highQC:     gNode,
			log:        log,
			orphanList: list.New(),
			orphanMap:  make(map[string]bool),
		}
	}
	// 重启状态时将root->tipBlock-3, generic->tipBlock-2, highQC->tipBlock-1
	// 若tipBlock<=2, root->genesisBlock, highQC->tipBlock-1
	tipNode := NewTreeNode(ledger, tip.GetHeight())
	if tip.GetHeight() < 3 {
		tree := &QCPendingTree{
			genesis:    gNode,
			root:       NewTreeNode(ledger, 0),
			log:        log,
			orphanList: list.New(),
			orphanMap:  make(map[string]bool),
		}
		switch tip.GetHeight() {
		case 0:
			tree.highQC = tree.root
			return tree
		case 1:
			tree.highQC = tree.root
			tree.highQC.Sons = append(tree.highQC.Sons, tipNode)
			return tree
		case 2:
			tree.highQC = NewTreeNode(ledger, 1)
			tree.highQC.Sons = append(tree.highQC.Sons, tipNode)
			tree.root.Sons = append(tree.root.Sons, tree.highQC)
		}
		return tree
	}
	tree := &QCPendingTree{
		genesis:    gNode,
		root:       NewTreeNode(ledger, tip.GetHeight()-3),
		genericQC:  NewTreeNode(ledger, tip.GetHeight()-2),
		highQC:     NewTreeNode(ledger, tip.GetHeight()-1),
		log:        log,
		orphanList: list.New(),
		orphanMap:  make(map[string]bool),
	}
	// 手动组装Tree结构
	tree.root.Sons = append(tree.root.Sons, tree.genericQC)
	tree.genericQC.Sons = append(tree.genericQC.Sons, tree.highQC)
	tree.highQC.Sons = append(tree.highQC.Sons, tipNode)
	return tree
}

func (t *QCPendingTree) MockGetOrphan() *list.List {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.orphanList
}

func (t *QCPendingTree) GetGenesisQC() *ProposalNode {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.genesis
}

func (t *QCPendingTree) GetRootQC() *ProposalNode {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.root
}

func (t *QCPendingTree) GetGenericQC() *ProposalNode {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.genericQC
}

func (t *QCPendingTree) GetCommitQC() *ProposalNode {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.commitQC
}

func (t *QCPendingTree) GetLockedQC() *ProposalNode {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.lockedQC
}

func (t *QCPendingTree) GetHighQC() *ProposalNode {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.highQC
}

// DFSQueryNode实现的比较简单，从root节点开始寻找，后续有更优方法可优化
func (t *QCPendingTree) DFSQueryNode(id []byte) *ProposalNode {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return dfsQuery(t.root, id)
}

// updateCommit 此方法向存储接口发送一个ProcessCommit，通知存储落盘，此时的block将不再被回滚
// 同时此方法将原先的root更改为commit node，因为commit node在本BFT中已确定不会回滚
func (t *QCPendingTree) UpdateCommit(id []byte) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	node := dfsQuery(t.root, id)
	if node == nil {
		return
	}
	parent := dfsQuery(t.root, node.In.GetParentProposalId())
	if parent == nil {
		return
	}
	parentParent := dfsQuery(t.root, parent.In.GetParentProposalId())
	if parentParent == nil {
		return
	}
	parentParentParent := dfsQuery(t.root, parentParent.In.GetParentProposalId())
	if parentParentParent == nil {
		return
	}
	parentParentParentParent := dfsQuery(t.root, parentParentParent.In.GetParentProposalId())
	if parentParentParentParent == nil {
		return
	}
	parentParentParentParent.Sons = nil
	t.root = parentParentParent
	// TODO: commitQC/lockedQC/genericQC/highQC是否有指向原root及以上的Node
}

// 更新本地qcTree, insert新节点, 将新节点parentQC和本地HighQC对比，如有必要进行更新
func (t *QCPendingTree) UpdateQcStatus(node *ProposalNode) error {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if node.Sons == nil {
		node.Sons = make([]*ProposalNode, 0)
	}
	if dfsQuery(t.root, node.In.GetProposalId()) != nil {
		logrus.Debug("QCPendingTree::updateQcStatus::has been inserted", "search", fmt.Sprintf("%x", node.In.GetProposalId()))
		return nil
	}
	if err := t.insert(node); err != nil {
		logrus.Error("QCPendingTree::updateQcStatus insert err", "err", err)
		return err
	}
	logrus.Debug("QCPendingTree::updateQcStatus", "insert new", fmt.Sprintf("%x", node.In.GetProposalId()), "height", node.In.GetProposalView(), "highQC", fmt.Sprintf("%x", t.highQC.In.GetProposalId()))

	// HighQCs试图更新成收到node的parentQC
	parent := dfsQuery(t.root, node.In.GetParentProposalId())
	if parent == nil {
		logrus.Debug("QCPendingTree::updateHighQC::orphan", "id", fmt.Sprintf("%x", node.In.GetParentProposalId()))
		return nil
	}
	// 若新验证过的node和原HighQC高度相同，使用新验证的node
	if parent.In.GetProposalView() < t.highQC.In.GetProposalView() {
		return nil
	}
	t.updateQCs(parent)
	return nil
}

// updateHighQC 对比QC树，将本地HighQC和输入id比较，高度更高的更新为HighQC，此时连同GenericQC、LockedQC、CommitQC一起修改
func (t *QCPendingTree) UpdateHighQC(inProposalId []byte) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	node := dfsQuery(t.root, inProposalId)
	if node == nil {
		logrus.Debug("QCPendingTree::updateHighQC::dfsQuery nil!", "id", fmt.Sprintf("%x", inProposalId))
		return
	}
	// 若新验证过的node和原HighQC高度相同，使用新验证的node
	if node.In.GetProposalView() < t.highQC.In.GetProposalView() {
		return
	}
	t.updateQCs(node)
}

// enforceUpdateHighQC 强制更改HighQC指针，用于错误时回滚，注意: 本实现没有timeoutQC因此需要此方法
func (t *QCPendingTree) EnforceUpdateHighQC(inProposalId []byte) error {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	node := dfsQuery(t.root, inProposalId)
	if node == nil {
		logrus.Debug("QCPendingTree::enforceUpdateHighQC::dfsQuery nil")
		return ErrNoValidQC
	}
	logrus.Debug("QCPendingTree::enforceUpdateHighQC::start.")
	return t.updateQCs(node)
}

func (t *QCPendingTree) updateQCs(highQCNode *ProposalNode) error {
	// 更改HighQC以及一系列的GenericQC、LockedQC和CommitQC
	t.highQC = highQCNode
	t.genericQC = nil
	t.lockedQC = nil
	t.commitQC = nil
	logrus.Debug("QCPendingTree::updateHighQC", "HighQC height", highQCNode.In.GetProposalView(), "HighQC", fmt.Sprintf("%x", highQCNode.In.GetProposalId()))
	parent := dfsQuery(t.root, highQCNode.In.GetParentProposalId())
	if parent == nil {
		return nil
	}
	t.genericQC = parent
	logrus.Debug("QCPendingTree::updateHighQC", "GenericQC height", t.genericQC.In.GetProposalView(), "GenericQC", fmt.Sprintf("%x", t.genericQC.In.GetProposalId()))
	// 找grand节点，标为LockedQC
	parentParent := dfsQuery(t.root, parent.In.GetParentProposalId())
	if parentParent == nil {
		return nil
	}
	t.lockedQC = parentParent
	logrus.Debug("QCPendingTree::updateHighQC", "LockedQC height", t.lockedQC.In.GetProposalView(), "LockedQC", fmt.Sprintf("%x", t.lockedQC.In.GetProposalId()))
	// 找grandgrand节点，标为CommitQC
	parentParentParent := dfsQuery(t.root, parentParent.In.GetParentProposalId())
	if parentParentParent == nil {
		return nil
	}
	t.commitQC = parentParentParent
	logrus.Debug("QCPendingTree::updateHighQC", "CommitQC height", t.commitQC.In.GetProposalView(), "CommitQC", fmt.Sprintf("%x", t.commitQC.In.GetProposalId()))
	return nil
}

// insert 向本地QC树Insert一个ProposalNode，如有必要，连同HighQC、GenericQC、LockedQC、CommitQC一起修改
func (t *QCPendingTree) insert(node *ProposalNode) error {
	if node.In == nil {
		logrus.Error("QCPendingTree::insert err", "err", ErrNoValidQC)
		return ErrNoValidQC
	}
	if node.In.GetParentProposalId() == nil {
		return ErrNoValidParentId
	}
	parent := dfsQuery(t.root, node.In.GetParentProposalId())
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
	if _, ok := t.orphanMap[fmt.Sprintf("%x", node.In.GetProposalId())]; ok {
		return nil // 重复退出
	}
	t.orphanMap[fmt.Sprintf("%x", node.In.GetProposalId())] = true
	if t.orphanList.Len() == 0 {
		t.orphanList.PushBack(node)
		return nil
	}
	// 遍历整个Sli，查看是否能够挂上
	ptr := t.orphanList.Front()
	for ptr != nil {
		curPtr := ptr
		n, ok := curPtr.Value.(*ProposalNode)
		if !ok {
			return errors.New("QCPendingTree::insertOrphan::element type invalid")
		}
		ptr = ptr.Next()
		// 查看头节点是否已经时间失效了, 失效的时候所有依赖该高度长得树实际上都没有意义了，需要删除
		if n.In.GetProposalView() <= t.root.In.GetProposalView() {
			t.orphanList.Remove(curPtr)
			continue
		}
		// 查看头节点是否是node的儿子, 直接在头部插入
		if bytes.Equal(n.In.GetParentProposalId(), node.In.GetProposalId()) {
			node.Sons = append(node.Sons, n)
			t.orphanList.Remove(curPtr)
			t.orphanList.PushBack(node)
			return nil
		}
		// 否则遍历该树试图挂在子树上面
		parent := dfsQuery(n, node.In.GetParentProposalId())
		if parent != nil {
			parent.Sons = append(parent.Sons, node)
			return nil
		}
	}
	// 没有可以挂的地方，则直接append
	t.orphanList.PushBack(node)
	return nil
}

// adoptOrphans 查看孤儿节点列表是否可以挂在该节点上
func (t *QCPendingTree) adoptOrphans(node *ProposalNode) error {
	if t.orphanList.Len() == 0 {
		return nil
	}
	ptr := t.orphanList.Front()
	for ptr != nil {
		curPtr := ptr
		n, ok := curPtr.Value.(*ProposalNode)
		if !ok {
			return errors.New("QCPendingTree::insertOrphan::element type invalid")
		}
		ptr = ptr.Next()
		if bytes.Equal(n.In.GetParentProposalId(), node.In.GetProposalId()) {
			node.Sons = append(node.Sons, n)
			t.orphanList.Remove(curPtr)
		}
	}
	return nil
}

func dfsQuery(node *ProposalNode, target []byte) *ProposalNode {
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
		if n := dfsQuery(node, target); n != nil {
			return n
		}
	}
	return nil
}
