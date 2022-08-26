package state

import (
	"container/list"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
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

type MptKV struct {
	// blockHash -> stateRoot
	indexDB KV

	nodeBase *NodeBase

	prevBlock      Hash
	currentBlock   Hash
	finalizedBlock Hash

	// FIXME: use ArrayList
	stashes *list.List // []*TxnStashes
}

const MptIndex = "mpt-index"

func NewMptKV(kvdb Kvdb) IState {

	nodeBase := NewNodeBase(kvdb)

	return &MptKV{
		indexDB:      kvdb.New(MptIndex),
		nodeBase:     nodeBase,
		prevBlock:    NullHash,
		currentBlock: NullHash,
		stashes:      list.New(),
	}
}

func (skv *MptKV) NextTxn() {
	skv.stashes.PushBack(newTxnStashes())
}

func (skv *MptKV) Set(triName NameString, key, value []byte) {
	skv.mute(SetOp, triName, key, value)
}

func (skv *MptKV) Delete(triName NameString, key []byte) {
	skv.mute(DeleteOp, triName, key, nil)
}

func (skv *MptKV) mute(op Ops, triName NameString, key, value []byte) {
	if skv.stashes.Len() == 0 {
		skv.stashes.PushBack(newTxnStashes())
	}
	skv.stashes.Back().Value.(*TxnStashes).append(op, makeKey(triName, key), value)
}

func (skv *MptKV) Get(triName NameString, key []byte) ([]byte, error) {
	for element := skv.stashes.Back(); element != nil; element = element.Prev() {
		stashes := element.Value.(*TxnStashes)
		ops, value := stashes.get(makeKey(triName, key))
		if ops != nil {
			if *ops == DeleteOp {
				return nil, nil
			}
			if value != nil {
				return value, nil
			}
		}
	}
	return skv.GetByBlockHash(triName, key, skv.prevBlock)
}

func (skv *MptKV) GetFinalized(triName NameString, key []byte) ([]byte, error) {
	return skv.GetByBlockHash(triName, key, skv.finalizedBlock)
}

func (skv *MptKV) Exist(triName NameString, key []byte) bool {
	value, _ := skv.Get(triName, key)
	return value != nil
}

func (skv *MptKV) GetByBlockHash(triName NameString, key []byte, blockHash Hash) ([]byte, error) {
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

// Commit returns StateRoot or error
func (skv *MptKV) Commit() (Hash, error) {
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
	for element := skv.stashes.Front(); element != nil; element = element.Next() {
		stashes := element.Value.(*TxnStashes)
		err = stashes.commit(mpt)
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

	skv.stashes.Init()
	return stateRoot, nil
}

func (skv *MptKV) Discard() {
	last := skv.stashes.Back()
	if last != nil {
		skv.stashes.Remove(last)
	}
}

func (skv *MptKV) DiscardAll() {
	stateRoot, err := skv.getIndexDB(skv.prevBlock)
	if err != nil {
		logrus.Panic("DiscardAll: get stateRoot error: ", err)
	}
	err = skv.setIndexDB(skv.currentBlock, stateRoot)
	if err != nil {
		logrus.Panic("DiscardAll: set stateRoot error: ", err)
	}

	skv.stashes.Init()
}

func (skv *MptKV) StartBlock(blockHash Hash) {
	skv.prevBlock = skv.currentBlock
	skv.currentBlock = blockHash
}

func (skv *MptKV) FinalizeBlock(blockHash Hash) {
	skv.finalizedBlock = blockHash
}

func (skv *MptKV) setIndexDB(blockHash, stateRoot Hash) error {
	return skv.indexDB.Set(blockHash.Bytes(), stateRoot.Bytes())
}

func (skv *MptKV) getIndexDB(blockHash Hash) (Hash, error) {
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
	stashes *list.List // []*KvStash
	// key: string(key bytes)
	indexes map[string]*list.Element
}

func newTxnStashes() *TxnStashes {
	return &TxnStashes{
		stashes: list.New(),
		indexes: make(map[string]*list.Element),
	}
}

func (k *TxnStashes) append(ops Ops, key, value []byte) {
	newKvStash := &KvStash{
		ops:   ops,
		Key:   key,
		Value: value,
	}
	last := k.stashes.PushBack(newKvStash)
	k.indexes[string(key)] = last
}

func (k *TxnStashes) get(key []byte) (*Ops, []byte) {
	if element, ok := k.indexes[string(key)]; ok {
		kvStash := element.Value.(*KvStash)
		return &kvStash.ops, kvStash.Value
	}
	return nil, nil
}

func (k *TxnStashes) commit(mpt *Trie) error {
	for element := k.stashes.Front(); element != nil; element = element.Next() {
		stash := element.Value.(*KvStash)
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
