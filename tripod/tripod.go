package tripod

import (
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/txn"
)

type Tripod interface {
	GetTripodMeta() *TripodMeta

	Name() string

	CheckTxn(*SignedTxn) error

	VerifyBlock(block IBlock, env *ChainEnv) bool

	InitChain(env *ChainEnv, land *Land) error

	StartBlock(block IBlock, env *ChainEnv, land *Land) error

	EndBlock(block IBlock, env *ChainEnv, land *Land) error

	FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error
}
