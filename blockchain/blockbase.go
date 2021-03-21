package blockchain

import (
	. "yu/common"
	"yu/event"
	ysql "yu/storage/sql"
	"yu/txn"
)

type BlockBase struct {
	db ysql.SqlDB
}

func NewBlockBase(db ysql.SqlDB) *BlockBase {
	return &BlockBase{
		db: db,
	}
}

func (bb *BlockBase) GetTxn(txnHash Hash) (txn.IsignedTxn, error) {

}

func (bb *BlockBase) SetTxn(txnHash Hash, stxn txn.IsignedTxn) error {

}

func (bb *BlockBase) GetTxns(blockHash Hash) ([]txn.IsignedTxn, error) {

}

func (bb *BlockBase) SetTxns(blockHash Hash, txns txn.SignedTxns) error {

}

func (bb *BlockBase) GetEvents(blockHash Hash) ([]event.IEvent, error) {

}

func (bb *BlockBase) SetEvents(blockHash Hash, events event.Events) error {

}
