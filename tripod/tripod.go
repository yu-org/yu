package tripod

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/chain_env"
	. "github.com/Lawliet-Chan/yu/txn"
)

type Tripod interface {
	GetTripodMeta() *TripodMeta

	Name() string

	CheckTxn(*SignedTxn) error

	ValidateBlock(block IBlock, env *ChainEnv) bool

	InitChain(env *ChainEnv, land *Land) error

	StartBlock(env *ChainEnv, land *Land) (newBlock IBlock, needBroadcast bool, err error)

	EndBlock(block IBlock, env *ChainEnv, land *Land) error

	FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error
}
