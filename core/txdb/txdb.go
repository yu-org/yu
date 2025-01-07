package txdb

import (
	"github.com/sirupsen/logrus"

	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/infra/storage/sql"
)

const (
	Txns    = "txns"
	Results = "results"
)

type TxDB struct {
	nodeType int

	txnKV     kv.KV
	receiptKV kv.KV

	enableUseSql bool
	db           sql.SqlDB
}

type TxnDBSchema struct {
	Type  string `gorm:"type:varchar(10)"`
	Key   string `gorm:"primaryKey;type:text"`
	Value string `gorm:"type:text"`
}

func (TxnDBSchema) TableName() string {
	return "txndb"
}

func NewTxDB(nodeTyp int, kvdb kv.Kvdb, kvdbConf *config.KVconf) (ItxDB, error) {
	txdb := &TxDB{
		nodeType:  nodeTyp,
		txnKV:     kvdb.New(Txns),
		receiptKV: kvdb.New(Results),
	}
	if kvdbConf != nil && kvdbConf.UseSQlDbConf {
		db, err := sql.NewSqlDB(&kvdbConf.SQLDbConf)
		if err != nil {
			return nil, err
		}
		txdb.db = db
		txdb.enableUseSql = true
		if err := txdb.db.AutoMigrate(&TxnDBSchema{}); err != nil {
			return nil, err
		}
	}
	return txdb, nil
}

func (bb *TxDB) GetTxn(txnHash Hash) (*SignedTxn, error) {
	if bb.nodeType == LightNode {
		return nil, nil
	}
	if bb.enableUseSql {
		var records []TxnDBSchema
		err := bb.db.Db().Raw("select value from txndb where type = ? and key = ?", "txn", string(txnHash.Bytes())).Find(&records).Error
		// find result in sql database
		if err == nil && len(records) > 0 {
			return DecodeSignedTxn([]byte(records[0].Value))
		}
	}
	byt, err := bb.txnKV.Get(txnHash.Bytes())
	if err != nil {
		return nil, err
	}
	if byt == nil {
		return nil, nil
	}
	return DecodeSignedTxn(byt)
}

func (bb *TxDB) GetTxns(txnHashes []Hash) ([]*SignedTxn, error) {
	if bb.nodeType == LightNode {
		return nil, nil
	}
	txns := make([]*SignedTxn, 0)
	for _, txnHash := range txnHashes {
		result, err := bb.GetTxn(txnHash)
		if err != nil {
			return nil, err
		}
		txns = append(txns, result)
	}
	return txns, nil
}

func (bb *TxDB) ExistTxn(txnHash Hash) bool {
	if bb.nodeType == LightNode {
		return false
	}
	if bb.enableUseSql {
		var records []TxnDBSchema
		err := bb.db.Db().Raw("select value from txndb where type = ? and key = ?", "txn", string(txnHash.Bytes())).Find(&records).Error
		if err == nil && len(records) > 0 {
			return true
		}
	}
	return bb.txnKV.Exist(txnHash.Bytes())
}

func (bb *TxDB) SetTxns(txns []*SignedTxn) error {
	if bb.nodeType == LightNode {
		return nil
	}
	if bb.enableUseSql {
		for _, txn := range txns {
			txbyt, err := txn.Encode()
			if err != nil {
				logrus.Errorf("TxDB.SetTxns set tx(%s) failed: %v", txn.TxnHash.String(), err)
				return err
			}
			if err := bb.db.Db().Exec("insert into txndb (type,key,value) values (?,?,?)", "txn", string(txn.TxnHash.Bytes()), string(txbyt)).Error; err != nil {
				logrus.Errorf("Insert TxDB.SetTxns tx(%s) failed: %v", txn.TxnHash.String(), err)
				return err
			}
		}
		return nil
	}
	kvtx, err := bb.txnKV.NewKvTxn()
	if err != nil {
		return err
	}
	for _, txn := range txns {
		txbyt, err := txn.Encode()
		if err != nil {
			logrus.Errorf("TxDB.SetTxns set tx(%s) failed: %v", txn.TxnHash.String(), err)
			return err
		}
		err = kvtx.Set(txn.TxnHash.Bytes(), txbyt)
		if err != nil {
			return err
		}
	}
	return kvtx.Commit()
}

func (bb *TxDB) SetReceipts(receipts map[Hash]*Receipt) error {
	if bb.enableUseSql {
		for txHash, receipt := range receipts {
			if err := bb.SetReceipt(txHash, receipt); err != nil {
				return err
			}
		}
		return nil
	}
	kvtx, err := bb.receiptKV.NewKvTxn()
	if err != nil {
		return err
	}
	for txHash, receipt := range receipts {
		byt, err := receipt.Encode()
		if err != nil {
			return err
		}
		err = kvtx.Set(txHash.Bytes(), byt)
		if err != nil {
			return err
		}
	}

	return kvtx.Commit()
}

func (bb *TxDB) SetReceipt(txHash Hash, receipt *Receipt) error {
	if bb.enableUseSql {
		byt, err := receipt.Encode()
		if err != nil {
			return err
		}
		if err := bb.db.Db().Exec("insert into txndb (type,key,value) values (?,?,?)", "receipt", string(txHash.Bytes()), string(byt)).Error; err != nil {
			return err
		}
		return nil
	}
	byt, err := receipt.Encode()
	if err != nil {
		return err
	}
	return bb.receiptKV.Set(txHash.Bytes(), byt)
}

func (bb *TxDB) GetReceipt(txHash Hash) (*Receipt, error) {
	if bb.enableUseSql {
		var records []TxnDBSchema
		err := bb.db.Db().Raw("select value from txndb where type = ? and key = ?", "receipt", string(txHash.Bytes())).Find(&records).Error
		if err == nil && len(records) > 0 {
			receipt := new(Receipt)
			err = receipt.Decode([]byte(records[0].Value))
			if err == nil {
				return receipt, nil
			}
		}
	}
	byt, err := bb.receiptKV.Get(txHash.Bytes())
	if err != nil {
		logrus.Errorf("TxDB.GetReceipt(%s), failed: %s, error: %v", txHash.String(), string(byt), err)
		return nil, err
	}
	if byt == nil {
		return nil, nil
	}
	receipt := new(Receipt)
	err = receipt.Decode(byt)
	if err != nil {
		logrus.Errorf("TxDB.GetReceipt(%s), Decode failed: %s, error: %v", txHash.String(), string(byt), err)
	}
	return receipt, err
}
