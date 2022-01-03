package state

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/infra/storage/kv"
	. "github.com/yu-org/yu/infra/trie/mpt"
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

	prevBlock      Hash
	currentBlock   Hash
	finalizedBlock Hash

	stashes []*TxnStashes
}

func NewStateKV(cfg *StateKvConf) IState {
	indexDB, err := NewKV(&cfg.IndexDB)
	if err != nil {
		logrus.Fatal("init stateKV indexDB error: ", err)
	}

	nodeBase, err := NewNodeBase(&cfg.NodeBase)
	if err != nil {
		logrus.Fatal("init stateKV nodeBase error: ", err)
	}

	return &StateKV{
		indexDB:      indexDB,
		nodeBase:     nodeBase,
		prevBlock:    NullHash,
		currentBlock: NullHash,
		stashes:      make([]*TxnStashes, 0),
	}
}

func (skv *StateKV) NextTxn() {
	skv.stashes = append(skv.stashes, newTxnStashes())
}

func (skv *StateKV) Set(triName NameString, key, value []byte) {
	skv.mute(SetOp, triName, key, value)
}

func (skv *StateKV) Delete(triName NameString, key []byte) {
	skv.mute(DeleteOp, triName, key, nil)
}

func (skv *StateKV) mute(op Ops, triName NameString, key, value []byte) {
	if len(skv.stashes) == 0 {
		skv.stashes = append(skv.stashes, newTxnStashes())
	}
	currentTxnStashes := skv.stashes[len(skv.stashes)-1]
	currentTxnStashes.append(op, makeKey(triName, key), value)
}

func (skv *StateKV) Get(triName NameString, key []byte) ([]byte, error) {
	for i := len(skv.stashes) - 1; i >= 0; i-- {
		value := skv.stashes[i].get(makeKey(triName, key))
		if value != nil {
			return value, nil
		}
	}
	return skv.GetByBlockHash(triName, key, skv.prevBlock)
}

func (skv *StateKV) GetFinalized(triName NameString, key []byte) ([]byte, error) {
	return skv.GetByBlockHash(triName, key, skv.finalizedBlock)
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
	lastStateRoot, err := skv.getIndexDB(skv.prevBlock)
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

	// todo: optimize combine all key-values stashes
	for _, stash := range skv.stashes {
		err = stash.commit(mpt)
		if err != nil {
			skv.DiscardAll()
			return NullHash, err
		}
	}

	stateRoot, err := mpt.Commit(nil)
	if err != nil {
		skv.DiscardAll()
		return NullHash, err
	}

	err = skv.setIndexDB(skv.currentBlock, stateRoot)
	if err != nil {
		skv.DiscardAll()
		return NullHash, err
	}

	skv.stashes = nil
	return stateRoot, nil
}

func (skv *StateKV) Discard() {
	if len(skv.stashes) == 0 {
		return
	}
	skv.stashes = skv.stashes[:len(skv.stashes)-1]
}

func (skv *StateKV) DiscardAll() {
	stateRoot, err := skv.getIndexDB(skv.prevBlock)
	if err != nil {
		logrus.Panic("DiscardAll: get stateRoot error: ", err)
	}
	err = skv.setIndexDB(skv.currentBlock, stateRoot)
	if err != nil {
		logrus.Panic("DiscardAll: set stateRoot error: ", err)
	}

	skv.stashes = nil
}

func (skv *StateKV) StartBlock(blockHash Hash) {
	skv.prevBlock = skv.currentBlock
	skv.currentBlock = blockHash
}

func (skv *StateKV) FinalizeBlock(blockHash Hash) {
	skv.finalizedBlock = blockHash
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

type TxnStashes struct {
	stashes []*KvStash
	// key: string(key bytes)
	// value: index of stashes
	indexes map[string]int
}

func newTxnStashes() *TxnStashes {
	return &TxnStashes{
		stashes: make([]*KvStash, 0),
		indexes: make(map[string]int),
	}
}

func (k *TxnStashes) append(ops Ops, key, value []byte) {
	newKvStash := &KvStash{
		ops:   ops,
		Key:   key,
		Value: value,
	}
	if idx, ok := k.indexes[string(key)]; ok {
		k.stashes = append(k.stashes[:idx], k.stashes[idx+1:]...)
	}
	k.stashes = append(k.stashes, newKvStash)
	k.indexes[string(key)] = len(k.stashes) - 1
}

func (k *TxnStashes) get(key []byte) []byte {
	if idx, ok := k.indexes[string(key)]; ok {
		return k.stashes[idx].Value
	}
	return nil
}

func (k *TxnStashes) commit(mpt *Trie) error {
	for _, stash := range k.stashes {
		switch stash.ops {
		case SetOp:
			err := mpt.TryUpdate(stash.Key, stash.Value)
			if err != nil {
				return err
			}
		case DeleteOp:
			err := mpt.TryDelete(stash.Key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type NameString interface {
	Name() string
}
