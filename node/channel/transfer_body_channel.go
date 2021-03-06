package channel

import (
	. "yu/node"
	. "yu/storage/queue"
)

type MemTbChan struct {
	c chan *TransferBody
}

func NewMemTbChan(cap int) TransferBodyChannel {
	return &MemTbChan{c: make(chan *TransferBody, cap)}
}

func (mtc *MemTbChan) Push(tb *TransferBody) error {
	mtc.c <- tb
	return nil
}

func (mtc *MemTbChan) Pop() (*TransferBody, error) {
	tb := <-mtc.c
	return tb, nil
}

type DiskTbChan struct {
	q Queue
}

func NewDiskTbChan(q Queue) TransferBodyChannel {
	return &DiskTbChan{
		q: q,
	}
}

func (dtc *DiskTbChan) Push(tb *TransferBody) error {
	byt, err := tb.Encode()
	if err != nil {
		return err
	}
	return dtc.q.Push(byt)
}

func (dtc *DiskTbChan) Pop() (*TransferBody, error) {
	byt, err := dtc.q.Pop()
	if err != nil {
		return nil, err
	}
	return DecodeTb(byt)
}
