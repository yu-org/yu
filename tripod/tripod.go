package tripod

import (
	. "yu/blockchain"
	"yu/txn"
	. "yu/txpool"
)

type Tripod interface {
	TripodMeta() *TripodMeta

	CheckTxn(*txn.SignedTxn) error

	ValidateBlock(IBlockChain, IBlock) bool

	InitChain(IBlockChain, IBlockBase) error

	StartBlock(IBlockChain, IBlock, ItxPool) (needBroadcast bool, err error)

	EndBlock(IBlockChain, IBlock, ItxPool) error

	FinalizeBlock(IBlockChain, IBlock) error
}
