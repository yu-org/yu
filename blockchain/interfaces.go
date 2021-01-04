package blockchain

import (
	. "yu/common"
	"yu/event"
	"yu/txn"
)

type IBlock interface {
	Head() IHeader
	Txns() []*txn.Txn
	Events() []event.Event
}

type IHeader interface {
	Num() BlockNum
	PreHash() Hash
	TxnRoot() Hash
	StateRoot() Hash
}
