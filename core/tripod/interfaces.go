package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

//type Tripod interface {
//	GetTripodHeader() *Tripod
//}

type BlockVerifier interface {
	VerifyBlock(block *Block) bool
}

type Init interface {
	InitChain(block *Block)
}

type BlockCycle interface {
	BlockStarter
	BlockEnder
	BlockFinalizer
}

type BlockStarter interface {
	StartBlock(block *Block)
}

type BlockEnder interface {
	EndBlock(block *Block)
}

type BlockFinalizer interface {
	FinalizeBlock(block *Block)
}

type Committer interface {
	Commit(ctx *Block)
}
