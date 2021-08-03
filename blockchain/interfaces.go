package blockchain

import (
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/result"
	. "github.com/Lawliet-Chan/yu/txn"
	"github.com/libp2p/go-libp2p-core/peer"
)

type IBlock interface {
	IHeader
	GetHeader() IHeader

	GetBlockId() BlockId
	GetTxnsHashes() []Hash
	SetTxnsHashes(hashes []Hash)

	SetHash(hash Hash)
	SetPreHash(hash Hash)
	SetTxnRoot(hash Hash)
	SetStateRoot(hash Hash)
	SetHeight(BlockNum)
	SetTimestamp(ts uint64)
	SetPeerID(peer.ID)

	SetLeiLimit(e uint64)
	UseLei(e uint64)

	Encode() ([]byte, error)
	Decode(data []byte) (IBlock, error)

	CopyFrom(other IBlock)
}

type IHeader interface {
	GetHeight() BlockNum
	GetHash() Hash
	GetPrevHash() Hash
	GetTxnRoot() Hash
	GetStateRoot() Hash
	GetTimestamp() uint64
	GetPeerID() peer.ID
	GetLeiLimit() uint64
	GetLeiUsed() uint64
}

// --------------- blockchain interface ----------------

type ConvergeType int

const (
	Longest ConvergeType = iota
	Heaviest
	Finalize
)

type IBlockChain interface {
	ConvergeType() ConvergeType

	NewEmptyBlock() IBlock

	EncodeBlocks(blocks []IBlock) ([]byte, error)
	DecodeBlocks(data []byte) ([]IBlock, error)

	// get genesis block
	GetGenesis() (IBlock, error)
	// set genesis block
	SetGenesis(b IBlock) error
	// pending a block from other blockchain-node for validating and operating
	InsertBlockFromP2P(ib IBlock) error

	TakeP2pBlocksBefore(height BlockNum) (map[BlockNum][]IBlock, error)
	TakeP2pBlocks(height BlockNum) ([]IBlock, error)

	AppendBlock(b IBlock) error
	GetBlock(blockHash Hash) (IBlock, error)
	ExistsBlock(blockHash Hash) bool
	UpdateBlock(b IBlock) error

	Children(prevBlockHash Hash) ([]IBlock, error)
	Finalize(blockHash Hash) error
	LastFinalized() (IBlock, error)
	GetEndBlock() (IBlock, error)
	GetAllBlocks() ([]IBlock, error)

	GetRangeBlocks(startHeight, endHeight BlockNum) ([]IBlock, error)

	Chain() (IChainStruct, error)
}

type IChainStruct interface {
	Append(block IBlock)
	InsertPrev(block IBlock)
	First() IBlock
	Range(fn func(block IBlock) error) error
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
	SetError(err *Error) error
}
