package state

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/storage/kv"
	. "github.com/yu-org/yu/trie/mpt"
)

//                         Merkle Patricia Trie
//                   /              |              \
//                  /               |               \
//  blockHash->stateRoot   blockHash->stateRoot  blockHash->stateRoot
//		/     |     \         /     |     \        /     |     \
//     /      |      \		 /      |      \      /      |      \
//	  kv      kv      kv     kv     kv     kv    kv     kv      kv

type StateKV struct {
	// blockHash -> stateRoot
	indexDB KV

	nodeBase *NodeBase

	nowBlock     Hash
	canReadBlock Hash

	nowStashes []*KvStash
	stashes    []*KvStash
}

func NewStateKV(cfg *StateKvConf) (*StateKV, error) {
	indexDB, err := NewKV(&cfg.IndexDB)
	if err != nil {
		return nil, err
	}

	nodeBase, err := NewNodeBase(&cfg.NodeBase)
	if err != nil {
		return nil, err
	}

	return &StateKV{
		indexDB:    indexDB,
		nodeBase:   nodeBase,
		nowBlock:   NullHash,
		nowStashes: make([]*KvStash, 0),
		stashes:    make([]*KvStash, 0),
	}, nil
}

func (skv *StateKV) NextTxn() {
	for _, stash := range skv.nowStashes {
		skv.stashes = append(skv.stashes, stash)
	}
	skv.nowStashes = make([]*KvStash, 0)
}

func (skv *StateKV) Set(triName NameString, key, value []byte) {
	skv.nowStashes = append(skv.nowStashes, &KvStash{
		ops:   SetOp,
		Key:   makeKey(triName, key),
		Value: value,
	})
}

func (skv *StateKV) Delete(triName NameString, key []byte) {
	skv.nowStashes = append(skv.nowStashes, &KvStash{
		ops:   DeleteOp,
		Key:   makeKey(triName, key),
		Value: nil,
	})
}

func (skv *StateKV) Get(triName NameString, key []byte) ([]byte, error) {
	return skv.GetByBlockHash(triName, key, skv.canReadBlock)
}

func (skv *StateKV) Exist(triName NameString, key []byte) bool {
	value, _ := skv.Get(triName, key)
	return value != nil
}

func (skv *StateKV) GetByBlockHash(triName NameString, key []byte, blockHash Hash) ([]byte, error) {
	stateRoot, err := skv.getIndexDB(blockHash)
	if err != nil {
		return nil, err
	}
	mpt, err := NewTrie(stateRoot, skv.nodeBase)
	if err != nil {
		return nil, err
	}
	return mpt.TryGet(makeKey(triName, key))
}

// return StateRoot or error
func (skv *StateKV) Commit() (Hash, error) {
	lastStateRoot, err := skv.getIndexDB(skv.canReadBlock)
	if err != nil {
		return NullHash, err
	}
	if lastStateRoot == NullHash {
		lastStateRoot = EmptyRoot
	}
	mpt, err := NewTrie(lastStateRoot, skv.nodeBase)
	if err != nil {
		skv.DiscardAll()
		return NullHash, err
	}
	for _, stash := range skv.stashes {
		switch stash.ops {
		case SetOp:
			err := mpt.TryUpdate(stash.Key, stash.Value)
			if err != nil {
				skv.DiscardAll()
				return NullHash, err
			}
		case DeleteOp:
			err := mpt.TryDelete(stash.Key)
			if err != nil {
				skv.DiscardAll()
				return NullHash, err
			}
		}
	}

	stateRoot, err := mpt.Commit(nil)
	if err != nil {
		skv.DiscardAll()
		return NullHash, err
	}

	err = skv.setIndexDB(skv.nowBlock, stateRoot)
	if err != nil {
		skv.DiscardAll()
		return NullHash, err
	}

	skv.stashes = nil
	return stateRoot, nil
}

func (skv *StateKV) Discard() {
	skv.nowStashes = nil
}

func (skv *StateKV) DiscardAll() {
	stateRoot, err := skv.getIndexDB(skv.canReadBlock)
	if err != nil {
		logrus.Panicf("DiscardAll: get stateRoot error: %s", err.Error())
	}
	err = skv.setIndexDB(skv.nowBlock, stateRoot)
	if err != nil {
		logrus.Panicf("DiscardAll: set stateRoot error: %s", err.Error())
	}

	skv.stashes = nil
}

func (skv *StateKV) StartBlock(blockHash Hash) {
	skv.nowBlock = blockHash
}

func (skv *StateKV) SetCanRead(blockHash Hash) {
	skv.canReadBlock = blockHash
}

func (skv *StateKV) setIndexDB(blockHash, stateRoot Hash) error {
	return skv.indexDB.Set(blockHash.Bytes(), stateRoot.Bytes())
}

func (skv *StateKV) getIndexDB(blockHash Hash) (Hash, error) {
	stateRoot, err := skv.indexDB.Get(blockHash.Bytes())
	if err != nil {
		return NullHash, err
	}
	return BytesToHash(stateRoot), nil
}

func makeKey(triName NameString, key []byte) []byte {
	tripodName := []byte(triName.Name())
	return append(tripodName, key...)
}

type Ops int

const (
	SetOp = iota
	DeleteOp
)

type KvStash struct {
	ops   Ops
	Key   []byte
	Value []byte
}

type NameString interface {
	Name() string
}
