package channel

import (
	. "yu/storage/queue"
	. "yu/txn"
	. "yu/utils/error_handle"
)

type MemTxnsChan struct {
	c chan IsignedTxn
}

func NewMemTxnsChan(cap int) TxnsChannel {
	return &MemTxnsChan{c: make(chan IsignedTxn, cap)}
}

func (mtc *MemTxnsChan) SendChan() chan<- IsignedTxn {
	return mtc.c
}

func (mtc *MemTxnsChan) RecvChan() <-chan IsignedTxn {
	return mtc.c
}

type DiskTxnsChan struct {
	snd  chan IsignedTxn
	recv chan IsignedTxn
	q    Queue
}

func NewDiskTxnsChan(cap int, q Queue) TxnsChannel {
	snd := make(chan IsignedTxn, cap)
	recv := make(chan IsignedTxn, cap)
	go LogfIfErr(q.PushAsync(TxnsTopic, snd), "push txns into queue error: ")
	go LogfIfErr(q.PopAsync(TxnsTopic, recv), "pop txns from queue error: ")

	return &DiskTxnsChan{
		snd:  snd,
		recv: recv,
		q:    q,
	}
}

func (dtc *DiskTxnsChan) SendChan() chan<- IsignedTxn {
	return dtc.snd
}

func (dtc *DiskTxnsChan) RecvChan() <-chan IsignedTxn {
	return dtc.recv
}
