package blockchain

import "github.com/yu-org/yu/types"

// todo: consider use array instead of linklist
type ChainStruct struct {
	//root *ChainNode
	chainArr []types.IBlock
}

func NewEmptyChain(block types.IBlock) *ChainStruct {
	return &ChainStruct{
		chainArr: []types.IBlock{block},
	}
}

func MakeFinalizedChain(blocks []types.IBlock) types.IChainStruct {
	chain := NewEmptyChain(blocks[0])
	for i := 1; i < len(blocks); i++ {
		chain.Append(blocks[i])
	}
	return chain
}

//func MakeLongestChain(blocks []IBlock) []IChainStruct {
//	longestChains := make([]IChainStruct, 0)
//	allChains := make([][]IBlock, 0)
//	for _, block := range blocks {
//		h := int(block.GetHeight())
//
//	}
//allBlocks := make(map[Hash]IBlock)
//
//highestBlocks := make([]IBlock, 0)
//
//var longestHeight BlockNum = 0
//for _, block := range blocks {
//	bh := block.GetHeight()
//	if bh > longestHeight {
//		longestHeight = bh
//		highestBlocks = nil
//	}
//
//	if bh == longestHeight {
//		highestBlocks = append(highestBlocks, block)
//	}
//
//	allBlocks[block.GetHash()] = block
//}
//
//for _, hblock := range highestBlocks {
//	chain := NewEmptyChain(hblock)
//	// FIXME: genesis block cannot be returned if its prevHash is Null
//	for chain.root.Current.GetPrevHash() != NullHash {
//		block, ok := allBlocks[chain.root.Current.GetPrevHash()]
//		if ok {
//			chain.InsertPrev(block)
//		}
//	}
//
//	longestChains = append(longestChains, chain)
//
//}

//	logrus.Warn("end RANGE highest blocks------------")
//
//	return longestChains
//}

// // deprecated
func MakeHeaviestChain(blocks []types.IBlock) []types.IChainStruct {
	return nil
}

func (c *ChainStruct) Append(block types.IBlock) {
	//cursor := c.root
	//for cursor.Next != nil {
	//	cursor = cursor.Next
	//}
	//cursor.Next = &ChainNode{
	//	Prev:    cursor,
	//	Current: block,
	//	Next:    nil,
	//}
	c.chainArr = append(c.chainArr, block)
}

func (c *ChainStruct) InsertPrev(block types.IBlock) {
	//c.root.Prev = &ChainNode{
	//	Prev:    nil,
	//	Current: block,
	//	Next:    c.root,
	//}
	//c.root = c.root.Prev
	c.chainArr = append([]types.IBlock{block}, c.chainArr...)
}

func (c *ChainStruct) First() types.IBlock {
	return c.chainArr[0]
}

func (c *ChainStruct) Range(fn func(block types.IBlock) error) error {
	//for cursor := c.root; cursor.Next != nil; cursor = cursor.Next {
	//	err := fn(cursor.Current)
	//	if err != nil {
	//		return err
	//	}
	//}
	for _, b := range c.chainArr {
		err := fn(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ChainStruct) Last() types.IBlock {
	//cursor := c.root
	//for cursor.Next != nil {
	//	cursor = cursor.Next
	//}
	//return cursor.Current
	return c.chainArr[len(c.chainArr)-1]
}

//type ChainNode struct {
//	Prev    *ChainNode
//	Current IBlock
//	Next    *ChainNode
//}
