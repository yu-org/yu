package blockchain

import (
	. "yu/common"
)

type IBlock interface {
	BlockId() BlockId
	TxnsHashes() []Hash
	SetTxnsHashes(hashes []Hash)

	SetPreHash(hash Hash)
	SetStateRoot(hash Hash)
	SetBlockNumber(BlockNum)

	Header() IHeader
	Extra() interface{}
	SetExtra(extra interface{})

	Encode() ([]byte, error)
	Decode(data []byte) (IBlock, error)
}

type IHeader interface {
	BlockNumber() BlockNum
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
