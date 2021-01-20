package mpt

import (
	"sync"
	. "yu/common"
	"yu/config"
	"yu/storage/kv"
)

type NodeBase struct {
	db   kv.KV
	lock sync.RWMutex
}

func NewNodeBase(cfg *config.KVconf) (*NodeBase, error) {
	db, err := kv.NewKV(cfg)
	if err != nil {
		return nil, err
	}
	return &NodeBase{db: db}, nil
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
