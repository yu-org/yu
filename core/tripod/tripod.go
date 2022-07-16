package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

type Tripod interface {
	GetTripodHeader() *TripodHeader

	InitChain()
}

type BlockVerifier interface {
	VerifyBlock(block *Block) bool
}

type BlockCycle interface {
	StartBlock(block *Block)
	EndBlock(block *Block)
	FinalizeBlock(block *Block)
}
