package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

type Tripod interface {
	GetTripodHeader() *TripodHeader

	CheckTxn(*SignedTxn) error
	VerifyBlock(block *CompactBlock) bool

	InitChain()
	StartBlock(block *CompactBlock)
	EndBlock(block *CompactBlock)
	FinalizeBlock(block *CompactBlock)
}
