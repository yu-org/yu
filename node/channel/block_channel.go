package channel

import (
	. "yu/blockchain"
	. "yu/storage/queue"
	. "yu/utils/error_handle"
)

type MemBlockChan struct {
	c chan IBlock
}

func NewMemBlockChan(cap int) BlockChannel {
	return &MemBlockChan{c: make(chan IBlock, cap)}
}

func (mbc *MemBlockChan) SendChan() chan<- IBlock {
	return mbc.c
}

func (mbc *MemBlockChan) RecvChan() <-chan IBlock {
	return mbc.c
}

type DiskBlockChan struct {
	snd  chan IBlock
	recv chan IBlock
	q    Queue
}

func NewDiskBlockChan(cap int, q Queue) BlockChannel {
	snd := make(chan IBlock, cap)
	recv := make(chan IBlock, cap)
	go LogfIfErr(q.PushAsync(BlockTopic, snd), "push block into queue error: ")
	go LogfIfErr(q.PopAsync(BlockTopic, recv), "pop block from queue error: ")

	return &DiskBlockChan{
		snd:  snd,
		recv: recv,
		q:    q,
	}
}

func (dbc *DiskBlockChan) SendChan() chan<- IBlock {
	return dbc.snd
}

func (dbc *DiskBlockChan) RecvChan() <-chan IBlock {
	return dbc.recv
}
