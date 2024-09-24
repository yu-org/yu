package state

import (
	"bytes"
	"container/list"
	"crypto/sha256"
	"github.com/celestiaorg/smt"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/metrics"
	"time"
)

//                         Merkle Patricia Trie
//                   /              |              \
//                  /               |               \
//  blockHash->stateRoot   blockHash->stateRoot  blockHash->stateRoot
//		/     |     \         /     |     \        /     |     \
//     /      |      \		 /      |      \      /      |      \
//	  kv      kv      kv     kv     kv     kv    kv     kv      kv

type SpmtKV struct {
	// blockHash -> stateRoot
	indexDB kv.KV

	// for spmt
	nodesDB  kv.KV
	valuesDB kv.KV

	spmt *smt.SparseMerkleTree

	prevBlock      *types.Block
	currentBlock   *types.Block
	finalizedBlock *types.Block

	// FIXME: use ArrayList
	stashes *list.List // []*TxnStashes
}

const (
	SpmtIndex = "spmt-index"
	Nodes     = "spmt-nodes"
	Values    = "spmt-values"
)

var (
	hasher    = sha256.New
	EmptyRoot = HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

func NewSpmtKV(root []byte, kvdb kv.Kvdb) IState {
	indexDB := kvdb.New(SpmtIndex)
	nodesDB := kvdb.New(Nodes)
	valuesDB := kvdb.New(Values)

	var spmt *smt.SparseMerkleTree
	if root == nil {
		spmt = smt.NewSparseMerkleTree(nodesDB, valuesDB, hasher())
	} else {
		spmt = smt.ImportSparseMerkleTree(nodesDB, valuesDB, hasher(), root)
	}

	return &SpmtKV{
		indexDB:      indexDB,
		nodesDB:      nodesDB,
		valuesDB:     valuesDB,
		spmt:         spmt,
		prevBlock:    nil,
		currentBlock: nil,
		stashes:      list.New(),
	}
}

func (skv *SpmtKV) NextTxn() {
	skv.stashes.PushBack(newTxnStashes())
}

func (skv *SpmtKV) Set(triName NameString, key, value []byte) {
	skv.set(triName.Name(), key, value)
}

func (skv *SpmtKV) set(triName string, key, value []byte) {
	skv.mute(SetOp, triName, key, value)
}

func (skv *SpmtKV) Delete(triName NameString, key []byte) {
	skv.delete(triName.Name(), key)
}

func (skv *SpmtKV) delete(triName string, key []byte) {
	skv.mute(DeleteOp, triName, key, nil)
}

func (skv *SpmtKV) mute(op Ops, triName string, key, value []byte) {
	if skv.stashes.Len() == 0 {
		skv.stashes.PushBack(newTxnStashes())
	}
	skv.stashes.Back().Value.(*TxnStashes).append(op, makeKey(triName, key), value)
}

func (skv *SpmtKV) Get(triName NameString, key []byte) ([]byte, error) {
	return skv.get(triName.Name(), key)
}

func (skv *SpmtKV) get(triName string, key []byte) ([]byte, error) {
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
	return skv.getByBlockHash(triName, key, skv.prevBlock)
}

func (skv *SpmtKV) GetFinalized(triName NameString, key []byte) ([]byte, error) {
	return skv.GetByBlockHash(triName, key, skv.finalizedBlock)
}

func (skv *SpmtKV) getFinalized(triName string, key []byte) ([]byte, error) {
	return skv.getByBlockHash(triName, key, skv.finalizedBlock)
}

func (skv *SpmtKV) Exist(triName NameString, key []byte) bool {
	return skv.exist(triName.Name(), key)
}

func (skv *SpmtKV) exist(triName string, key []byte) bool {
	value, _ := skv.get(triName, key)
	return value != nil
}

// FIXME
func (skv *SpmtKV) GetByBlockHash(triName NameString, key []byte, block *types.Block) ([]byte, error) {
	return skv.getByBlockHash(triName.Name(), key, block)
}

func (skv *SpmtKV) getByBlockHash(triName string, key []byte, block *types.Block) ([]byte, error) {
	key = makeKey(triName, key)
	//stateRoot, err := skv.getIndexDB(blockHash)
	//if err != nil {
	//	return nil, err
	//}

	// mpt := smt.ImportSparseMerkleTree(skv.nodesDB, skv.valuesDB, hasher(), stateRoot)
	value, err := skv.spmt.Get(key)
	if bytes.Equal(value, []byte{}) {
		// because of https://github.com/celestiaorg/smt/blob/master/smt.go#L14
		value = nil
	}
	return value, err
}

// Commit returns StateRoot or error
func (skv *SpmtKV) Commit() ([]byte, error) {
	//lastStateRoot, err := skv.getIndexDB(skv.prevBlock)
	//if err != nil {
	//	return nil, err
	//}
	//if lastStateRoot == nil {
	//	lastStateRoot = EmptyRoot.Bytes()
	//}

	//spmt := smt.NewSparseMerkleTree(skv.nodesDB, skv.valuesDB, hasher())

	start := time.Now()
	defer func() {
		metrics.StateCommitDuration.WithLabelValues().Observe(time.Since(start).Seconds())
	}()

	// todo: optimize combine all key-values stashes
	for element := skv.stashes.Front(); element != nil; element = element.Next() {
		stashes := element.Value.(*TxnStashes)
		err := stashes.apply(skv.spmt)
		if err != nil {
			skv.DiscardAll()
			return nil, err
		}
	}

	//stateRoot, err := spmt.Commit(nil)
	//if err != nil {
	//	skv.DiscardAll()
	//	return NullHash, err
	//}
	stateRoot := skv.spmt.Root()

	err := skv.setIndexDB(skv.currentBlock, stateRoot)
	if err != nil {
		skv.DiscardAll()
		return nil, err
	}

	skv.stashes.Init()
	return stateRoot, nil
}

func (skv *SpmtKV) Discard() {
	last := skv.stashes.Back()
	if last != nil {
		skv.stashes.Remove(last)
	}
}

func (skv *SpmtKV) DiscardAll() {
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

func (skv *SpmtKV) StartBlock(block *types.Block) {
	skv.prevBlock = skv.currentBlock
	skv.currentBlock = block
}

func (skv *SpmtKV) FinalizeBlock(block *types.Block) {
	skv.finalizedBlock = block
}

func (skv *SpmtKV) setIndexDB(block *types.Block, stateRoot []byte) error {
	return skv.indexDB.Set(block.Hash.Bytes(), stateRoot)
}

func (skv *SpmtKV) getIndexDB(block *types.Block) ([]byte, error) {
	stateRoot, err := skv.indexDB.Get(block.Hash.Bytes())
	if err != nil {
		return nil, err
	}
	return stateRoot, nil
}

func makeKey(triName string, key []byte) []byte {
	tripodName := []byte(triName)
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

func (k *TxnStashes) apply(mpt *smt.SparseMerkleTree) error {
	for element := k.stashes.Front(); element != nil; element = element.Next() {
		stash := element.Value.(*KvStash)
		switch stash.ops {
		case SetOp:
			_, err := mpt.Update(stash.Key, stash.Value)
			if err != nil {
				return err
			}
		case DeleteOp:
			_, err := mpt.Delete(stash.Key)
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
