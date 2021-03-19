package blockchain

type ChainStruct struct {
	root *ChainNode
}

func NewEmptyChain(root IBlock) *ChainStruct {
	return &ChainStruct{
		root: &ChainNode{
			Current: root,
			Next:    nil,
		},
	}
}

func MakeLongestChain(root IBlock, blocks []IBlock) *ChainStruct {
	cs := NewEmptyChain(root)
	cursor := cs.root.Current
	chains := make([]*ChainStruct, 0)
	for _, block := range blocks {
		bh := block.Header()
		ch := cursor.Header()
		if bh.PrevHash() == ch.Hash() {

		}
	}

	return cs
}

func MakeHeaviestChain(root IBlock, blocks []IBlock) *ChainStruct {
	cs := NewEmptyChain(root)
	height := root.Header().Height() + 1
	for _, block := range blocks {

	}
}

func (c *ChainStruct) First() IBlock {
	return c.root.Current
}

func (c *ChainStruct) Last() IBlock {
	cursor := c.root
	for cursor.Next != nil {
		cursor = cursor.Next
	}
	return cursor.Current
}

type ChainNode struct {
	Current IBlock
	Next    *ChainNode
}
