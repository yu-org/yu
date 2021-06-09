package state

import (
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/config"
	. "github.com/Lawliet-Chan/yu/storage/kv"
	. "github.com/Lawliet-Chan/yu/trie/mpt"
	"github.com/sirupsen/logrus"
)

//                         Merkle Patricia Trie
//                   /              |              \
//                  /               |               \
//  blockHash, stateRoot   blockHash, stateRoot    blockHash, stateRoot
//		/     |     \         /     |     \        /     |     \
//     /      |      \		 /      |      \      /      |      \
//	  kv      kv      kv     kv     kv     kv    kv     kv      kv

type StateKV struct {
	// blockHash -> stateRoot
	indexDB KV

	nodeBase *NodeBase

	nowBlock     Hash
	canReadBlock Hash
	stashes      []*KvStash
}

func NewStateKV(cfg *StateKvConf, canReadBlock Hash) (*StateKV, error) {
	indexDB, err := NewKV(&cfg.IndexDB)
	if err != nil {
		return nil, err
	}

	nodeBase, err := NewNodeBase(&cfg.NodeBase)
	if err != nil {
		return nil, err
	}

	return &StateKV{
		indexDB:      indexDB,
		nodeBase:     nodeBase,
		nowBlock:     NullHash,
		canReadBlock: canReadBlock,
		stashes:      make([]*KvStash, 0),
	}, nil
}

func (skv *StateKV) Set(key, value []byte) {
	skv.stashes = append(skv.stashes, &KvStash{
		ops:   SetOp,
		Key:   key,
		Value: value,
	})
}

func (skv *StateKV) Delete(key []byte) {
	skv.stashes = append(skv.stashes, &KvStash{
		ops:   DeleteOp,
		Key:   key,
		Value: nil,
	})
}

func (skv *StateKV) Get(key []byte) ([]byte, error) {
	return skv.GetByBlockHash(key, skv.canReadBlock)
}

func (skv *StateKV) GetByBlockHash(key []byte, blockHash Hash) ([]byte, error) {
	stateRoot, err := skv.getIndexDB(blockHash)
	if err != nil {
		return nil, err
	}
	mpt, err := NewTrie(stateRoot, skv.nodeBase)
	if err != nil {
		return nil, err
	}
	return mpt.TryGet(key)
}

// return StateRoot or error
func (skv *StateKV) Commit() (Hash, error) {
	mpt, err := NewTrie(skv.nowBlock, skv.nodeBase)
	if err != nil {
		skv.Discard()
		return NullHash, err
	}
	for _, stash := range skv.stashes {
		switch stash.ops {
		case SetOp:
			err := mpt.TryUpdate(stash.Key, stash.Value)
			if err != nil {
				skv.Discard()
				return NullHash, err
			}
		case DeleteOp:
			err := mpt.TryDelete(stash.Key)
			if err != nil {
				skv.Discard()
				return NullHash, err
			}
		}
	}

	stateRoot, err := mpt.Commit(nil)
	if err != nil {
		skv.Discard()
		return NullHash, err
	}

	err = skv.setIndexDB(skv.nowBlock, stateRoot)
	if err != nil {
		skv.Discard()
		return NullHash, err
	}

	skv.flushStashes()
	return stateRoot, nil
}

func (skv *StateKV) Discard() {
	stateRoot, err := skv.getIndexDB(skv.canReadBlock)
	if err != nil {
		logrus.Panicf("Discard: get stateRoot error: %s", err.Error())
	}
	err = skv.setIndexDB(skv.nowBlock, stateRoot)
	if err != nil {
		logrus.Panicf("Discard: set stateRoot error: %s", err.Error())
	}

	skv.flushStashes()
}

func (skv *StateKV) flushStashes() {
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
