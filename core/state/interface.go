package state

import (
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
)

// TODO: need to add prove() and verify()
type IState interface {
	Set(triName NameString, key, value []byte)
	Delete(triName NameString, key []byte)
	Get(triName NameString, key []byte) ([]byte, error)
	GetFinalized(triName NameString, key []byte) ([]byte, error)
	Exist(triName NameString, key []byte) bool
	GetByBlockHash(triName NameString, key []byte, block *types.Block) ([]byte, error)
	Commit() ([]byte, error)
	NextTxn()
	Discard()
	DiscardAll()
	StartBlock(block *types.Block)
	FinalizeBlock(block *types.Block)
}

func NewStateDB(typ string, kvdb kv.Kvdb) IState {
	// FIXME
	switch typ {
	case "no":
		return new(NoStateDB)
	default:
		return NewSpmtKV(nil, kvdb)
	}
}
