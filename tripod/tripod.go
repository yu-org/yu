package tripod

import (
	. "github.com/yu-altar/yu/blockchain"
	. "github.com/yu-altar/yu/chain_env"
	. "github.com/yu-altar/yu/txn"
)

type Tripod interface {
	GetTripodMeta() *TripodMeta

	Name() string

	CheckTxn(*SignedTxn) error

	VerifyBlock(block IBlock, env *ChainEnv) bool

	InitChain(env *ChainEnv, land *Land) error

	StartBlock(block IBlock, env *ChainEnv, land *Land) (needBroadcast bool, err error)

	EndBlock(block IBlock, env *ChainEnv, land *Land) error

	FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error
}
