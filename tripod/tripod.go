package tripod

import (
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/txn"
)

type Tripod interface {
	GetTripodMeta() *TripodMeta

	Name() string

	SetChainEnv(env *ChainEnv)

	CheckTxn(*SignedTxn) error

	VerifyBlock(block IBlock) bool

	InitChain(land *Land) error

	StartBlock(block IBlock, land *Land) error

	EndBlock(block IBlock, land *Land) error

	FinalizeBlock(block IBlock, land *Land) error
}
