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

	InitChain() error

	StartBlock(block IBlock) error

	EndBlock(block IBlock) error

	FinalizeBlock(block IBlock) error
}
