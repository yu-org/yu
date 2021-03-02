package run

import (
	. "yu/blockchain"
	. "yu/tripod"
	"yu/txpool"
)

func Run(chain IBlockChain, land *Land, pool txpool.ItxPool) error {
	// todo: Do parent block must be last finalized block ?
	block, err := chain.LastFinalized()
	if err != nil {
		return err
	}

	height := block.BlockNumber()
	preHash := block.PrevHash()

	newBlock := NewBlock(preHash, height)

	// start a new block
	err = land.RangeList(func(tri Tripod) error {
		return tri.StartBlock(newBlock)
	})

	// execute these txns
	err = land.RangeList(func(tri Tripod) error {
		return tri.HandleTxns(newBlock, pool)
	})

	// end block and append to chain
	err = land.RangeList(func(tri Tripod) error {
		return tri.EndBlock(newBlock)
	})
	err = chain.AppendBlock(block)
	if err != nil {
		return err
	}

	// finalize this block
	err = land.RangeList(func(tri Tripod) error {
		return tri.FinalizeBlock(newBlock)
	})
	return chain.Finalize(newBlock.BlockId())
}
