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
	SetHash(hash Hash)
	StateRoot() Hash
	SetStateRoot(hash Hash)
	PrevHash() Hash
	Header() IHeader
	Txns() []*txn.Txn
	Events() []event.Event
	Timestamp() int64
	Extra() interface{}
	SetExtra(extra interface{})
}

type IHeader interface {
	Num() BlockNum
	PrevHash() Hash
	TxnRoot() Hash
	StateRoot() Hash
	Timestamp() int64
	Extra() interface{}
}

type IBlockChain interface {
	AppendBlock(b IBlock) error
	GetBlock(id BlockId) (IBlock, error)
	Children(id BlockId) ([]IBlock, error)
	Finalize(id BlockId) error
	LastFinalized() (IBlock, error)
	Leaves() ([]IBlock, error)
}
