package tripod

import (
	. "github.com/yu-org/yu/core/chain_env"
	types2 "github.com/yu-org/yu/core/types"
)

type Tripod interface {
	GetTripodMeta() *TripodMeta

	Name() string

	SetChainEnv(env *ChainEnv)

	CheckTxn(*types2.SignedTxn) error

	VerifyBlock(block *types2.CompactBlock) bool

	InitChain() error

	StartBlock(block *types2.CompactBlock) error

	EndBlock(block *types2.CompactBlock) error

	FinalizeBlock(block *types2.CompactBlock) error
}
