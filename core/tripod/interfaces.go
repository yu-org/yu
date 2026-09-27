package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

//type Tripod interface {
//	GetTripodHeader() *Tripod
//}

type BlockVerifier interface {
	VerifyBlock(block *Block) error
}

type Init interface {
	InitChain()
}

// GenesisDefiner is implemented by the tripod defining the genesis block. The kernel calls
// DefineGenesis once, before any InitChain, only when the chain has no genesis block yet.
type GenesisDefiner interface {
	DefineGenesis() *Block
}

type BlockCycle interface {
	StartBlock(block *Block)
	EndBlock(block *Block)
	FinalizeBlock(block *Block)
}

type Committer interface {
	Commit(ctx *Block)
}

type PreTxnHandler interface {
	PreHandleTxn(*SignedTxn) error
}
