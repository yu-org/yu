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
	InitChain(block *Block)
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
