package run

import (
	. "yu/blockchain"
	. "yu/tripod"
	"yu/txpool"
)

func Run(chain IBlockChain, land *Land, pool txpool.ItxPool) error {
	newBlock := NewDefaultBlock()

	// start a new block
	err := land.RangeList(func(tri Tripod) error {
		return tri.StartBlock(chain, newBlock)
	})

	// execute these txns
	err = land.RangeList(func(tri Tripod) error {
		return tri.HandleTxns(newBlock, pool)
	})

	// end block and append to chain
	err = land.RangeList(func(tri Tripod) error {
		return tri.EndBlock(newBlock)
	})
	err = chain.AppendBlock(newBlock)
	if err != nil {
		return err
	}

	// finalize this block
	return land.RangeList(func(tri Tripod) error {
		return tri.FinalizeBlock(chain, newBlock)
	})
}
