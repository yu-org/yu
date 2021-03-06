package queue

import "yu/storage"

type ChanQueue struct {
	ChanQueuene chan []byte
}

func NewChanQueue(cap int) *ChanQueue {
	return &ChanQueue{ChanQueuene: make(chan []byte, cap)}
}

func (*ChanQueue) Type() storage.StoreType {
	return storage.Embedded
}

func (*ChanQueue) Kind() storage.StoreKind {
	return storage.Queue
}

func (cq *ChanQueue) Push(e []byte) error {
	cq.ChanQueuene <- e
	return nil
}

func (cq *ChanQueue) Pop() ([]byte, error) {
	e := <-cq.ChanQueuene
	return e, nil
}
