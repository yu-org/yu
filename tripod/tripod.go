package tripod

import (
	. "yu/blockchain"
	"yu/txn"
	"yu/txpool"
)

type Tripod interface {
	TripodMeta() *TripodMeta

	CheckTxn(txn.IsignedTxn) error

	// ValidateBlock(IBlock) error

	StartBlock(IBlockChain, IBlock, txpool.ItxPool) error

	EndBlock(IBlockChain, IBlock) error

	FinalizeBlock(IBlockChain, IBlock) error
}
