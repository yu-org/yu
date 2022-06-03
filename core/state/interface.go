package state

import (
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/storage/kv"
)

type IState interface {
	NextTxn()
	Set(triName NameString, key, value []byte)
	Delete(triName NameString, key []byte)
	Get(triName NameString, key []byte) ([]byte, error)
	GetFinalized(triName NameString, key []byte) ([]byte, error)
	Exist(triName NameString, key []byte) bool
	GetByBlockHash(triName NameString, key []byte, blockHash Hash) ([]byte, error)
	Commit() (Hash, error)
	Discard()
	DiscardAll()
	StartBlock(blockHash Hash)
	FinalizeBlock(blockHash Hash)
}

func NewStateDB(kvdb kv.KV) IState {
	return NewMptKV(kvdb)
}
