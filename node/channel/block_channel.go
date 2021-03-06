package channel

import (
	. "yu/blockchain"
	. "yu/storage/queue"
)

type MemBlockChan struct {
	c chan IBlock
}

func NewMemBlockChan(cap int) BlockChannel {
	return &MemBlockChan{c: make(chan IBlock, cap)}
}

func (mbc *MemBlockChan) Push(block IBlock) error {
	mbc.c <- block
	return nil
}

func (mbc *MemBlockChan) Pop() (IBlock, error) {
	block := <-mbc.c
	return block, nil
}

type DiskBlockChan struct {
	q Queue
}

func NewDiskBlockChan(q Queue) BlockChannel {
	return &DiskBlockChan{
		q: q,
	}
}

func (dbc *DiskBlockChan) Push(block IBlock) error {
	byt, err := block.Encode()
	if err != nil {
		return err
	}
	return dbc.q.Push(byt)
}

func (dbc *DiskBlockChan) Pop() (IBlock, error) {
	block := NewEmptyBlock()
	byt, err := dbc.q.Pop()
	if err != nil {
		return nil, err
	}
	return block.Decode(byt)
}
