package blockchain

import (
	"time"
	"yu/event"
	"yu/txn"
)

type Block struct {
	header *Header
	txns []*txn.Txn

	ReceiveTime time.Time
}

func NewBlock(header *Header, txns []*txn.Txn) *Block {
	return &Block{
		header: header,
		txns: txns,
		ReceiveTime: time.Now(),
	}
}

func(b *Block) Head() IHeader {
	return b.header
}

func(b *Block) Txns() []*txn.Txn {
	return b.txns
}

func(b *Block) Events() []event.Event {
	allEvents := make([]event.Event, 0)
	for _, tx := range b.txns {
		events:= tx.Events()
		allEvents = append(allEvents, events...)
	}
	return allEvents
}