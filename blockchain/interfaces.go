package blockchain

import (
	. "yu/common"
	"yu/event"
	"yu/txn"
)

type IBlock interface {
	BlockId() BlockId
	BlockNumber() BlockNum
	Hash() Hash
	Header() IHeader
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
	Children(id BlockId) ([]IBlock, error)
	LastFinalized() (IBlock, error)
	Leaves() ([]IBlock, error)
}
