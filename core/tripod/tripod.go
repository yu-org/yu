package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

type Tripod interface {
	GetTripodHeader() *TripodHeader

	CheckTxn(*SignedTxn) error
	VerifyBlock(block *Block) bool

	InitChain()
	StartBlock(block *Block)
	EndBlock(block *Block)
	FinalizeBlock(block *Block)
}
