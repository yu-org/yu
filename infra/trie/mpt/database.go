package mpt

import (
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/storage/kv"
	"sync"
)

type NodeBase struct {
	db   kv.KV
	lock sync.RWMutex
}

const MptData = "mpt-data"

func NewNodeBase(db kv.Kvdb) *NodeBase {
	return &NodeBase{db: db.NewKVInstance(MptData)}
}

func (db *NodeBase) node(hash Hash) node {
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

func (db *NodeBase) Put(key []byte, value []byte) error {
	return db.db.Set(key, value)
}

func (db *NodeBase) Close() error {
	return nil
}

func (db *NodeBase) insert(hash Hash, blob []byte) {
	// fmt.Println("inserting", hash, blob)
	db.Put(hash.Bytes(), blob)
}
