package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

type Tripod interface {
	GetTripodHeader() *TripodHeader

	CheckTxn(*SignedTxn) error
	VerifyBlock(block *CompactBlock) bool

	InitChain() error
	StartBlock(block *CompactBlock) error
	EndBlock(block *CompactBlock) error
	FinalizeBlock(block *CompactBlock) error
}
