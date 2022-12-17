package mpt

import (
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/storage/kv"
	"sync"
)

type NodeBase struct {
	db   kv.KV
	tx   kv.KvTxn
	lock sync.RWMutex
}

const MptData = "mpt-data"

func NewNodeBase(db kv.Kvdb) *NodeBase {
	return &NodeBase{db: db.New(MptData)}
}

func (db *NodeBase) node(hash common.Hash) node {
	enc, err := db.db.Get(hash.Bytes())
	if err != nil || enc == nil {
		return nil
	}
	// fmt.Println("node", hex.EncodeToString(hash[:]) , "->", hex.EncodeToString(enc))
	return mustDecodeNode(hash.Bytes(), enc)
}

func (db *NodeBase) Get(toGet []byte) ([]byte, error) {
	return db.db.Get(toGet)
}

func (db *NodeBase) Close() error {
	return nil
}

func (db *NodeBase) Begin() (err error) {
	db.tx, err = db.db.NewKvTxn()
	return
}

func (db *NodeBase) Insert(hash common.Hash, blob []byte) error {
	return db.tx.Set(hash.Bytes(), blob)
}

func (db *NodeBase) Commit() error {
	return db.tx.Commit()
}
