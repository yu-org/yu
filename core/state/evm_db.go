package state

import (
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"math/big"
)

type EvmDB struct {
	DB *gstate.StateDB
}

func NewEvmDB(root Hash) IState {
	ethdb, err := rawdb.NewLevelDBDatabase("evmdb", 0, 0, "", false)
	if err != nil {
		logrus.Fatal("init eth rawdb error: ", err)
	}
	db, err := gstate.New(gcommon.Hash(root), gstate.NewDatabase(ethdb), nil)
	if err != nil {
		logrus.Fatal("init evm statedb error: ", err)
	}
	return &EvmDB{DB: db}
}

func (db *EvmDB) AddBalance(addr Address, b *big.Int) {
	db.DB.AddBalance(gcommon.Address(addr), b)
}

func (db *EvmDB) SubBalance(addr Address, b *big.Int) {
	db.DB.SubBalance(gcommon.Address(addr), b)
}

func (db *EvmDB) GetBalance(addr Address) *big.Int {
	return db.DB.GetBalance(gcommon.Address(addr))
}

func (db *EvmDB) NextTxn() {
	panic("implement me")
}

func (db *EvmDB) Set(triName NameString, key, value []byte) {
	panic("implement me")
}

func (db *EvmDB) Delete(triName NameString, key []byte) {
	panic("implement me")
}

func (db *EvmDB) Get(triName NameString, key []byte) ([]byte, error) {
	panic("implement me")
}

func (db *EvmDB) GetFinalized(triName NameString, key []byte) ([]byte, error) {
	panic("implement me")
}

func (db *EvmDB) Exist(triName NameString, key []byte) bool {
	panic("implement me")
}

func (db *EvmDB) GetByBlockHash(triName NameString, key []byte, blockHash Hash) ([]byte, error) {
	panic("implement me")
}

func (db *EvmDB) Commit() (Hash, error) {
	panic("implement me")
}

func (db *EvmDB) Discard() {
	panic("implement me")
}

func (db *EvmDB) DiscardAll() {
	panic("implement me")
}

func (db *EvmDB) StartBlock(blockHash Hash) {
	panic("implement me")
}

func (db *EvmDB) FinalizeBlock(blockHash Hash) {
	panic("implement me")
}
