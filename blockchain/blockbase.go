package blockchain

import (
	. "yu/common"
	. "yu/result"
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
	var stxn txn.IsignedTxn
	bb.db.Db().Where(&TxnContent{TxnHash: txnHash.String()}).First(&stxn)
	return stxn, nil
}

func (bb *BlockBase) SetTxn(stxn txn.IsignedTxn) error {
	txnCt, err := toTxnContent(stxn)
	if err != nil {
		return err
	}
	bb.db.Db().Create(&txnCt)
	return nil
}

func (bb *BlockBase) GetTxns(blockHash Hash) ([]txn.IsignedTxn, error) {
	var txns []txn.IsignedTxn
	bb.db.Db().Where(&TxnContent{BlockHash: blockHash.String()}).Find(&txns)
	return txns, nil
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

func (bb *BlockBase) GetEvents(blockHash Hash) ([]Event, error) {

}

func (bb *BlockBase) SetEvents(blockHash Hash, events []Event) error {

}

type TxnContent struct {
	TxnHash   string `gorm:"txn_hash;primaryKey"`
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
	return TxnContent{
		TxnHash:   stxn.GetTxnHash().String(),
		Pubkey:    stxn.GetPubkey().String(),
		Signature: ToHex(stxn.GetSignature()),
		RawTxn:    ToHex(rawTxnByt),
		BlockHash: "",
	}, nil
}
