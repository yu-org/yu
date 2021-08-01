// Copyright Xuperchain Authors
// link: https://github.com/xuperchain/xupercore

package hotstuff

type IProposerElection interface {
	// 获取指定round的主节点Address, 注意, 若存在validators变更, 则需要在此处进行addrToIntAddr的更新操作
	GetLeader(round int64) string
	// 获取指定round的候选人节点Address
	GetValidators(round int64) []string
	// 获取候选人地址到网络地址的映射, for unit test
	GetIntAddress(string) string
}
