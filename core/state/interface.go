package state

import (
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/storage/kv"
)

// TODO: need to add prove() and verify()
type IState interface {
	Set(triName NameString, key, value []byte)
	Delete(triName NameString, key []byte)
	Get(triName NameString, key []byte) ([]byte, error)
	GetFinalized(triName NameString, key []byte) ([]byte, error)
	Exist(triName NameString, key []byte) bool
	GetByBlockHash(triName NameString, key []byte, blockHash Hash) ([]byte, error)
	Commit() ([]byte, error)
	NextTxn()
	Discard()
	DiscardAll()
	StartBlock(blockHash Hash)
	FinalizeBlock(blockHash Hash)
}

func NewStateDB(kvdb kv.Kvdb) IState {
	return NewSpmtKV(kvdb)
}
