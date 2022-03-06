package tripod

import (
	. "github.com/yu-org/yu/core/chain_env"
	. "github.com/yu-org/yu/core/types"
)

type Tripod interface {
	GetTripodMeta() *TripodMeta

	Name() string

	SetLand(land *Land)

	SetChainEnv(env *ChainEnv)

	CheckTxn(*SignedTxn) error

	VerifyBlock(block *CompactBlock) bool

	InitChain() error

	StartBlock(block *CompactBlock) error

	EndBlock(block *CompactBlock) error

	FinalizeBlock(block *CompactBlock) error
}
