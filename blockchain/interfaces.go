package blockchain

import (
	"yu/event"
	"yu/txn"
)

type IBlock interface {
	Head()  IHeader
	Txns() []*txn.Txn
	Events() []event.Event
}

type IHeader interface {
	Num() BlockNum
	ParentHash() string
	TRoot() string
	SRoot() string
}

type BlockNum = uint64
