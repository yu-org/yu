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

	VerifyBlock(block types.IBlock) bool

	InitChain() error

	StartBlock(block types.IBlock) error

	EndBlock(block types.IBlock) error

	FinalizeBlock(block types.IBlock) error
}
