package state

import (
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/infra/storage/kv"
	. "github.com/yu-org/yu/infra/trie/mpt"
	"math/big"
)

type EvmDB struct {
	// blockHash -> stateRoot
	indexDB KV

	nodeBase *NodeBase

	prevBlock      Hash
	currentBlock   Hash
	finalizedBlock Hash

	DB *gstate.StateDB

	stashes []*EvmTxnStashes
}

func NewEvmDB(root Hash, cfg *StateEvmConf) IState {
	ethdb, err := rawdb.NewLevelDBDatabase(cfg.Fpath, cfg.Cache, cfg.Handles, cfg.Namespace, cfg.ReadOnly)
	if err != nil {
		logrus.Fatal("init geth rawdb error: ", err)
	}
	db, err := gstate.New(gcommon.Hash(root), gstate.NewDatabase(ethdb), nil)
	if err != nil {
		logrus.Fatal("init geth statedb error: ", err)
	}
	indexDB, err := NewKV(&cfg.IndexDB)
	if err != nil {
		logrus.Fatal("init EvmDB indexDB error: ", err)
	}

	nodeBase, err := NewNodeBase(&cfg.NodeBase)
	if err != nil {
		logrus.Fatal("init EvmDB nodeBase error: ", err)
	}

	return &EvmDB{
		DB:           db,
		indexDB:      indexDB,
		nodeBase:     nodeBase,
		prevBlock:    NullHash,
		currentBlock: NullHash,
		stashes:      make([]*EvmTxnStashes, 0),
	}
}

func (db *EvmDB) AddBalance(addr Address, b *big.Int) {
	db.DB.AddBalance(gcommon.Address(addr), b)
	db.muteBalance(AddBalance, addr, b)
}

func (db *EvmDB) SubBalance(addr Address, b *big.Int) {
	db.DB.SubBalance(gcommon.Address(addr), b)
	db.muteBalance(SubBalance, addr, b)
}

func (db *EvmDB) GetBalance(addr Address) *big.Int {
	return db.DB.GetBalance(gcommon.Address(addr))
}

func (db *EvmDB) muteBalance(op Ops, addr Address, b *big.Int) {
	db.mute().appendBalanceOp(op, addr, b)
}

func (db *EvmDB) muteKV(op Ops, triName NameString, key, value []byte) {
	db.mute().append(op, makeKey(triName, key), value)
}

func (db *EvmDB) mute() *EvmTxnStashes {
	if len(db.stashes) == 0 {
		db.stashes = append(db.stashes, newEvmTxnStashes())
	}
	return db.stashes[len(db.stashes)-1]
}

func (db *EvmDB) NextTxn() {
	db.stashes = append(db.stashes, newEvmTxnStashes())
}

func (db *EvmDB) Set(triName NameString, key, value []byte) {
	db.muteKV(SetOp, triName, key, value)
}

func (db *EvmDB) Delete(triName NameString, key []byte) {
	db.muteKV(DeleteOp, triName, key, nil)
}

func (db *EvmDB) Get(triName NameString, key []byte) ([]byte, error) {
	for i := len(db.stashes) - 1; i >= 0; i-- {
		value := db.stashes[i].get(makeKey(triName, key))
		if value != nil {
			return value, nil
		}
	}
	return db.GetByBlockHash(triName, key, db.prevBlock)
}

func (db *EvmDB) GetFinalized(triName NameString, key []byte) ([]byte, error) {
	return db.GetByBlockHash(triName, key, db.finalizedBlock)
}

func (db *EvmDB) Exist(triName NameString, key []byte) bool {
	value, _ := db.Get(triName, key)
	return value != nil
}

func (db *EvmDB) GetByBlockHash(triName NameString, key []byte, blockHash Hash) ([]byte, error) {
	stateRoot, err := db.getIndexDB(blockHash)
	if err != nil {
		return nil, err
	}
	mpt, err := NewTrie(stateRoot, db.nodeBase)
	if err != nil {
		return nil, err
	}
	return mpt.TryGet(makeKey(triName, key))
}

func (db *EvmDB) Commit() (Hash, error) {
	lastStateRoot, err := db.getIndexDB(db.prevBlock)
	if err != nil {
		return NullHash, err
	}
	if lastStateRoot == NullHash {
		lastStateRoot = EmptyRoot
	}
	mpt, err := NewTrie(lastStateRoot, db.nodeBase)
	if err != nil {
		db.DiscardAll()
		return NullHash, err
	}

	// todo: optimize combine all key-values stashes
	for _, stash := range db.stashes {
		err = stash.commit(mpt)
		if err != nil {
			db.DiscardAll()
			return NullHash, err
		}
	}

	stateRoot, err := mpt.Commit(nil)
	if err != nil {
		db.DiscardAll()
		return NullHash, err
	}

	err = db.setIndexDB(db.currentBlock, stateRoot)
	if err != nil {
		db.DiscardAll()
		return NullHash, err
	}

	db.stashes = nil
	return stateRoot, nil
}

func (db *EvmDB) Discard() {
	db.discardBalanceOps(db.stashes[len(db.stashes)-1].stashes)
	db.stashes = db.stashes[:len(db.stashes)-1]
}

func (db *EvmDB) DiscardAll() {
	stateRoot, err := db.getIndexDB(db.prevBlock)
	if err != nil {
		logrus.Panicf("DiscardAll: get stateRoot error: %s", err.Error())
	}
	err = db.setIndexDB(db.currentBlock, stateRoot)
	if err != nil {
		logrus.Panicf("DiscardAll: set stateRoot error: %s", err.Error())
	}

	allStashes := make([]*EvmKvStash, 0)
	for _, stash := range db.stashes {
		for _, kvStash := range stash.stashes {
			allStashes = append(allStashes, kvStash)
		}
	}
	db.discardBalanceOps(allStashes)

	db.stashes = nil
}

func (db *EvmDB) discardBalanceOps(stashes []*EvmKvStash) {
	for _, stash := range stashes {
		if isBalanceOp(stash.ops) {
			if stash.ops == AddBalance {
				db.DB.SubBalance(gcommon.Address(stash.addr), stash.amount)
			}
			if stash.ops == SubBalance {
				db.DB.AddBalance(gcommon.Address(stash.addr), stash.amount)
			}
		}
	}
}

func (db *EvmDB) StartBlock(blockHash Hash) {
	db.prevBlock = db.currentBlock
	db.currentBlock = blockHash
}

func (db *EvmDB) FinalizeBlock(blockHash Hash) {
	db.finalizedBlock = blockHash
}

func (db *EvmDB) setIndexDB(blockHash, stateRoot Hash) error {
	return db.indexDB.Set(blockHash.Bytes(), stateRoot.Bytes())
}

func (db *EvmDB) getIndexDB(blockHash Hash) (Hash, error) {
	stateRoot, err := db.indexDB.Get(blockHash.Bytes())
	if err != nil {
		return NullHash, err
	}
	return BytesToHash(stateRoot), nil
}

const (
	AddBalance Ops = 3
	SubBalance Ops = 4
)

func isBalanceOp(op Ops) bool {
	return op == AddBalance || op == SubBalance
}

type EvmKvStash struct {
	*KvStash
	addr   Address
	amount *big.Int
}

type EvmTxnStashes struct {
	stashes []*EvmKvStash
	// key: string(address or bytes)
	// value: index of stashes
	indexes map[string]int
}

func newEvmTxnStashes() *EvmTxnStashes {
	return &EvmTxnStashes{
		stashes: make([]*EvmKvStash, 0),
		indexes: make(map[string]int),
	}
}

func (e *EvmTxnStashes) get(key []byte) []byte {
	if idx, ok := e.indexes[string(key)]; ok {
		return e.stashes[idx].Value
	}
	return nil
}

func (e *EvmTxnStashes) append(op Ops, key, value []byte) {
	newStash := &EvmKvStash{
		KvStash: &KvStash{
			ops:   op,
			Key:   key,
			Value: value,
		},
		addr:   Address{},
		amount: nil,
	}
	if idx, ok := e.indexes[string(key)]; ok {
		e.stashes = append(e.stashes[:idx], e.stashes[idx+1:]...)
	}
	e.stashes = append(e.stashes, newStash)
	e.indexes[string(key)] = len(e.stashes) - 1
}

func (e *EvmTxnStashes) appendBalanceOp(ops Ops, addr Address, b *big.Int) {
	newEvmKvStash := &EvmKvStash{
		KvStash: &KvStash{
			ops: ops,
		},
		addr:   addr,
		amount: b,
	}
	if idx, ok := e.indexes[addr.String()]; ok {
		e.stashes = append(e.stashes[:idx], e.stashes[idx+1:]...)
	}
	e.stashes = append(e.stashes, newEvmKvStash)
	e.indexes[addr.String()] = len(e.stashes) - 1
}

func (e *EvmTxnStashes) commit(mpt *Trie) error {
	for _, stash := range e.stashes {
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
