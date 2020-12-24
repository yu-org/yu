package blockchain

import (
	"bytes"
	"encoding/gob"
	"time"
	"yu/event"
	"yu/txn"
)

type Block struct {
	header *Header
	txns   []*txn.Txn

	ReceiveTime time.Time
}

func NewBlock(header *Header, txns []*txn.Txn) *Block {
	return &Block{
		header:      header,
		txns:        txns,
		ReceiveTime: time.Now(),
	}
}

func (b *Block) Head() *Header {
	return b.header
}

func (b *Block) Txns() []*txn.Txn {
	return b.txns
}

func (b *Block) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte) (*Block, error) {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (b *Block) Events() []event.Event {
	allEvents := make([]event.Event, 0)
	for _, tx := range b.txns {
		events := tx.Events()
		allEvents = append(allEvents, events...)
	}
	return allEvents
}
