package blockchain

import (
	"github.com/HyperService-Consortium/go-hexutil"
	. "yu/common"
	"yu/event"
	ysql "yu/storage/sql"
	"yu/txn"
)

type BlockBase struct {
	db ysql.SqlDB
}

func NewBlockBase(db ysql.SqlDB) *BlockBase {
	db.Db().Create(&TxnContent{})
	return &BlockBase{
		db: db,
	}
}

func (bb *BlockBase) GetTxn(txnHash Hash) (txn.IsignedTxn, error) {
	txnHashStr := string(txnHash.Bytes())

}

func (bb *BlockBase) SetTxn(txnHash Hash, stxn txn.IsignedTxn) error {
	txnHashStr := string(txnHash.Bytes())
	rawTxn := stxn.GetRaw()
	rawTxnByt, err := rawTxn.Encode()
	if err != nil {
		return err
	}
	rawTxnStr := string(rawTxnByt)

}

func (bb *BlockBase) GetTxns(blockHash Hash) ([]txn.IsignedTxn, error) {
	bb.db.Db().Model(&TxnContent{BlockHash: blockHash.String()}).Find()
}

func (bb *BlockBase) SetTxns(blockHash Hash, txns []txn.IsignedTxn) error {
	txnCts := make([]TxnContent, 0)
	for _, stxn := range txns {
		txnCt, err := newTxnContent(blockHash, stxn)
		if err != nil {
			return err
		}
		txnCts = append(txnCts, txnCt)
	}
	bb.db.Db().Create(&txnCts)
	return nil
}

func (bb *BlockBase) GetEvents(blockHash Hash) ([]event.IEvent, error) {

}

func (bb *BlockBase) SetEvents(blockHash Hash, events event.Events) error {

}

type TxnContent struct {
	TxnHash   string `gorm:"txn_hash"`
	Pubkey    string `gorm:"pubkey"`
	Signature string `gorm:"signature"`
	RawTxn    string `gorm:"raw_txn"`

	BlockHash string `gorm:"block_hash"`
}

func newTxnContent(blockHash Hash, stxn txn.IsignedTxn) (TxnContent, error) {
	txnCt, err := toTxnContent(stxn)
	if err != nil {
		return TxnContent{}, err
	}
	txnCt.BlockHash = blockHash.String()
	return txnCt, nil
}

func toTxnContent(stxn txn.IsignedTxn) (TxnContent, error) {
	rawTxnByt, err := stxn.GetRaw().Encode()
	if err != nil {
		return TxnContent{}, err
	}
	rawTxn := hexutil.Encode(rawTxnByt)
	return TxnContent{
		TxnHash:   stxn.GetTxnHash().String(),
		Pubkey:    stxn.GetPubkey().String(),
		Signature: hexutil.Encode(stxn.GetSignature()),
		RawTxn:    rawTxn,
		BlockHash: "",
	}, nil
}
