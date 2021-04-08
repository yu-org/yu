package blockchain

import (
	. "yu/common"
	. "yu/result"
	. "yu/txn"
)

type IBlock interface {
	GetBlockId() BlockId
	GetTxnsHashes() []Hash
	SetTxnsHashes(hashes []Hash)

	SetHash(hash Hash)
	SetPreHash(hash Hash)
	SetStateRoot(hash Hash)
	SetHeight(BlockNum)

	GetHeader() IHeader
	GetExtra() interface{}
	SetExtra(extra interface{})

	Encode() ([]byte, error)
	Decode(data []byte) (IBlock, error)
}

type IHeader interface {
	GetHeight() BlockNum
	GetHash() Hash
	GetPrevHash() Hash
	GetTxnRoot() Hash
	GetStateRoot() Hash
	GetTimestamp() int64
	GetExtra() interface{}
}

type IBlockChain interface {
	NewEmptyBlock() IBlock
	// just generate a block with timestamp
	NewDefaultBlock() IBlock
	// set genesis block
	SetGenesis(b IBlock) error
	// pending a block from other blockchain-node for validating and operating
	InsertBlockFromP2P(ib IBlock) error
	// get a pending block
	GetBlocksFromP2P(height BlockNum) ([]IBlock, error)
	// remove a block when finished validating and operate the block
	FlushBlocksFromP2P(height BlockNum) error

	AppendBlock(b IBlock) error
	GetBlock(blockHash Hash) (IBlock, error)
	Children(prevBlockHash Hash) ([]IBlock, error)
	Finalize(blockHash Hash) error
	LastFinalized() (IBlock, error)
	Leaves() ([]IBlock, error)

	// return the longest children chains
	Longest() ([]IChainStruct, error)
	// return the heaviest children chains
	Heaviest() ([]IChainStruct, error)
}

type IChainStruct interface {
	Append(block IBlock)
	InsertPrev(block IBlock)
	First() IBlock
	Last() IBlock
}

type IBlockBase interface {
	GetTxn(txnHash Hash) (*SignedTxn, error)
	SetTxn(stxn *SignedTxn) error

	GetTxns(blockHash Hash) ([]*SignedTxn, error)
	SetTxns(blockHash Hash, txns []*SignedTxn) error

	GetEvents(blockHash Hash) ([]*Event, error)
	SetEvents(events []*Event) error

	GetErrors(blockHash Hash) ([]*Error, error)
	SetErrors(errs []*Error) error
}
