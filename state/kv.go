package state

import (
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/config"
	. "github.com/Lawliet-Chan/yu/trie/mpt"
)

//                         Merkle Patricia Trie
//                   /              |              \
//                  /               |               \
//        blockHash             blockHash            blockHash
//		/     |     \         /     |     \        /     |     \
//     /      |      \		 /      |      \      /      |      \
//	  kv      kv      kv     kv     kv     kv    kv     kv      kv

type StateKV struct {
	nodeBase *NodeBase

	nowBlock     Hash
	canReadBlock Hash
	stashes      []*KvStash
}

func NewStateKV(kvCfg *KVconf, canReadBlock Hash) (*StateKV, error) {
	nodeBase, err := NewNodeBase(kvCfg)
	if err != nil {
		return nil, err
	}

	return &StateKV{
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
	mpt, err := NewTrie(blockHash, skv.nodeBase)
	if err != nil {
		return nil, err
	}
	return mpt.TryGet(key)
}

func (skv *StateKV) Commit() error {
	defer skv.Discard()
	for _, stash := range skv.stashes {
		switch stash.ops {
		case SetOp:
			err := skv.insertTrie(stash.Key, stash.Value)
			if err != nil {
				return err
			}
		case DeleteOp:
			err := skv.removeTrie(stash.Key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (skv *StateKV) Discard() {
	skv.stashes = nil
}

func (skv *StateKV) insertTrie(key, value []byte) error {
	mpt, err := NewTrie(skv.nowBlock, skv.nodeBase)
	if err != nil {
		return err
	}
	return mpt.TryUpdate(key, value)
}

func (skv *StateKV) removeTrie(key []byte) error {
	mpt, err := NewTrie(skv.nowBlock, skv.nodeBase)
	if err != nil {
		return err
	}
	return mpt.TryDelete(key)
}

func (skv *StateKV) StartBlock(blockHash Hash) {
	skv.nowBlock = blockHash
}

func (skv *StateKV) SetCanRead(blockHash Hash) {
	skv.canReadBlock = blockHash
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
