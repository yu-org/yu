package tripod

import (
	. "yu/blockchain"
	. "yu/chain_env"
	. "yu/txn"
)

type Tripod interface {
	TripodMeta() *TripodMeta

	CheckTxn(*SignedTxn) error

	ValidateBlock(block IBlock, env *ChainEnv) bool

	InitChain(env *ChainEnv, land *Land) error

	StartBlock(env *ChainEnv, land *Land) (newBlock IBlock, needBroadcast bool, err error)

	EndBlock(block IBlock, env *ChainEnv, land *Land) error

	FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error
}
