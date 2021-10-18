package tripod

import (
	. "github.com/yu-org/yu/chain_env"
	"github.com/yu-org/yu/types"
)

type Tripod interface {
	GetTripodMeta() *TripodMeta

	Name() string

	SetChainEnv(env *ChainEnv)

	CheckTxn(*types.SignedTxn) error

	VerifyBlock(block *types.CompactBlock) bool

	InitChain() error

	StartBlock(block *types.CompactBlock) error

	EndBlock(block *types.CompactBlock) error

	FinalizeBlock(block *types.CompactBlock) error
}
