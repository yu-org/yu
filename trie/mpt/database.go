package mpt

import (
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	. "yu/common"
)

type NodeBase struct {
	handler *leveldb.DB
	lock sync.RWMutex
}

func NewNodeBasefromDB(db *leveldb.DB) (*NodeBase, error) {
	return &NodeBase{handler: db}, nil
}

func NewNodeBase(path string) (*NodeBase, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &NodeBase{handler: db}, nil
}

func (db *NodeBase) node(hash Hash) node {
	enc, err := db.handler.Get(hash[:], nil)
	if err != nil || enc == nil {
		return nil
	}
	// fmt.Println("node", hex.EncodeToString(hash[:]) , "->", hex.EncodeToString(enc))
	return mustDecodeNode(hash[:], enc)
}


func (db *NodeBase) Get(toGet []byte) ([]byte, error) {
	return db.handler.Get(toGet, nil)
}

func (db *NodeBase) Put(key []byte, value []byte) error {
	return db.handler.Put(key, value, nil)
}

func (db *NodeBase) Close() error {
	return db.handler.Close()
}

func (db *NodeBase) insert(hash Hash, blob []byte) {
	// fmt.Println("inserting", hash, blob)
	db.Put(hash.Bytes(), blob)
}