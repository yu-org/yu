package blockchain

import (
	. "yu/common"
	"yu/event"
	"yu/txn"
)

type IBlock interface {
	Txns() []*txn.Txn
	Events() []event.Event
}

type IHeader interface {
	Num() BlockNum
	PrevHash() Hash
	TxnRoot() Hash
	StateRoot() Hash
	Extra() interface{}
}

type IBlockChain interface {
	AppendBlock(b IBlock) error
	Children(hash Hash) []IBlock
	LastFinalized() IBlock
	Leaves() []IBlock
}
