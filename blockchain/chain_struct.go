package blockchain

import (
	. "yu/common"
)

// todo: consider use array instead of linklist
type ChainStruct struct {
	root *ChainNode
}

func NewEmptyChain(root IBlock) *ChainStruct {
	return &ChainStruct{
		root: &ChainNode{
			Prev:    nil,
			Current: root,
			Next:    nil,
		},
	}
}

func MakeLongestChain(blocks []IBlock) []IChainStruct {
	longestChains := make([]IChainStruct, 0)
	allBlocks := make(map[Hash]IBlock)

	highestBlocks := make([]IBlock, 0)

	var longestHeight BlockNum = 0
	for _, block := range blocks {
		bh := block.GetHeader().GetHeight()
		if bh > longestHeight {
			longestHeight = bh
			highestBlocks = nil
		}

		if bh == longestHeight {
			highestBlocks = append(highestBlocks, block)
		}

		allBlocks[block.GetHeader().GetHash()] = block
	}

	for _, hblock := range highestBlocks {
		chain := NewEmptyChain(hblock)
		// FIXME: genesis block cannot be returned if its prevHash is Null
		for chain.root.Current.GetHeader().GetPrevHash() != NullHash {
			block, ok := allBlocks[chain.root.Current.GetHeader().GetPrevHash()]
			if ok {
				chain.InsertPrev(block)
			}
		}

		longestChains = append(longestChains, chain)
	}

	return longestChains
}

// todo
func MakeHeaviestChain(blocks []IBlock) []IChainStruct {
	return nil
}

func (c *ChainStruct) Append(block IBlock) {
	cursor := c.root
	for cursor.Next != nil {
		cursor = cursor.Next
	}
	cursor.Next = &ChainNode{
		Prev:    cursor,
		Current: block,
		Next:    nil,
	}
}

func (c *ChainStruct) InsertPrev(block IBlock) {
	c.root.Prev = &ChainNode{
		Prev:    nil,
		Current: block,
		Next:    c.root,
	}
	c.root = c.root.Prev
}

func (c *ChainStruct) First() IBlock {
	return c.root.Current
}

func (c *ChainStruct) Range(fn func(block IBlock) error) error {
	for cursor := c.root; cursor.Next != nil; cursor = cursor.Next {
		err := fn(cursor.Current)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ChainStruct) Last() IBlock {
	cursor := c.root
	for cursor.Next != nil {
		cursor = cursor.Next
	}
	return cursor.Current
}

type ChainNode struct {
	Prev    *ChainNode
	Current IBlock
	Next    *ChainNode
}
