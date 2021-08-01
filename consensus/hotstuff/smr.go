// Copyright Xuperchain Authors
// link: https://github.com/xuperchain/xupercore

package hotstuff

import (
	"container/list"
	"sync"
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

	// quitCh stop channel
	QuitCh chan bool

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
