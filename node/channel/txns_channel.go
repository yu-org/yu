package channel

import (
	. "yu/storage/queue"
	. "yu/txn"
)

type MemTxnsChan struct {
	c chan IsignedTxn
}

func NewMemTxnsChan(cap int) TxnsChannel {
	return &MemTxnsChan{c: make(chan SignedTxns, cap)}
}

func (mtc *MemTxnsChan) Push(txn IsignedTxn) error {
	mtc.c <- txn
	return nil
}

func (mtc *MemTxnsChan) Pop(num int) (SignedTxns, error) {
	var txns SignedTxns
	for i := 0; i < num; i++ {
		txn := <-mtc.c
		txns = append(txns, txn)
	}
	return txns, nil
}

type DiskTxnsChan struct {
	q Queue
}

func NewDiskTxnsChan(q Queue) TxnsChannel {
	return &DiskTxnsChan{
		q: q,
	}
}

func (dtc *DiskTxnsChan) Push(txn IsignedTxn) error {
	byt, err := txn.Encode()
	if err != nil {
		return err
	}
	return dtc.q.Push(byt)
}

func (dtc *DiskTxnsChan) Pop(num int) (SignedTxns, error) {
	var txns SignedTxns
	for i := 0; i < num; i++ {
		byt, err := dtc.q.Pop()
		if err != nil {
			return nil, err
		}
		txn, err := DecodeSignedTxn(byt)
		if err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}
	return txns, nil
}
