package tripod

import (
	. "yu/blockchain"
	. "yu/env"
	. "yu/txn"
)

type Tripod interface {
	TripodMeta() *TripodMeta

	CheckTxn(*SignedTxn) error

	ValidateBlock(block IBlock, env *Env) bool

	InitChain(env *Env, land *Land) error

	StartBlock(env *Env, land *Land) (newBlock IBlock, needBroadcast bool, err error)

	EndBlock(block IBlock, env *Env, land *Land) error

	FinalizeBlock(block IBlock, env *Env, land *Land) error
}
