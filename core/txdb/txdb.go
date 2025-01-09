package txdb

import (
	"sync"

	"github.com/sirupsen/logrus"

	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/infra/storage/sql"
	"github.com/yu-org/yu/metrics"
)

const (
	Txns    = "txns"
	Results = "results"
)

type TxDB struct {
	nodeType int

	txnKV     *txnkvdb
	receiptKV *receipttxnkvdb

	enableUseSql bool
	db           sql.SqlDB
}

type txnkvdb struct {
	sync.RWMutex
	txnKV kv.KV
}

const (
	sqlSourceType = "sql"
	kvSourceType  = "kv"
	txnType       = "txn"
	receiptType   = "receipt"
	successStatus = "success"
	errStatus     = "err"
)

func getSourceTypeValue(enableSQL bool) string {
	if enableSQL {
		return sqlSourceType
	}
	return kvSourceType
}

func getStatusValue(err error) string {
	if err == nil {
		return successStatus
	}
	return errStatus
}

func (t *txnkvdb) GetTxn(txnHash Hash) (txn *SignedTxn, err error) {
	t.RLock()
	defer t.RUnlock()
	byt, err := t.txnKV.Get(txnHash.Bytes())
	if err != nil {
		return nil, err
	}
	if byt == nil {
		return nil, nil
	}
	return DecodeSignedTxn(byt)
}

func (t *txnkvdb) ExistTxn(txnHash Hash) bool {
	t.RLock()
	defer t.RUnlock()
	return t.txnKV.Exist(txnHash.Bytes())
}

func (t *txnkvdb) SetTxns(txns []*SignedTxn) (err error) {
	t.Lock()
	defer t.Unlock()
	kvtx, err := t.txnKV.NewKvTxn()
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

type TxnDBSchema struct {
	Type    string `gorm:"type:varchar(10)"`
	HashKey string `gorm:"primaryKey,length:255;type:text"`
	Value   []byte `gorm:"type:mediumblob"`
}

func (TxnDBSchema) TableName() string {
	return "txndb"
}

func NewTxDB(nodeTyp int, kvdb kv.Kvdb, kvdbConf *config.KVconf) (ItxDB, error) {
	txdb := &TxDB{
		nodeType:  nodeTyp,
		txnKV:     &txnkvdb{txnKV: kvdb.New(Txns)},
		receiptKV: &receipttxnkvdb{receiptKV: kvdb.New(Results)},
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

func (bb *TxDB) GetTxn(txnHash Hash) (stxn *SignedTxn, err error) {
	if bb.nodeType == LightNode {
		return nil, nil
	}
	r, err := bb.txnKV.GetTxn(txnHash)
	if err == nil && r == nil {
		var records []TxnDBSchema
		err := bb.db.Db().Raw("select value from txndb where type = ? and hash_key = ?", "txn", txnHash.String()).Find(&records).Error
		// find result in sql database
		if err == nil && len(records) > 0 {
			res, err := DecodeSignedTxn(records[0].Value)
			metrics.TxnDBCounter.WithLabelValues(txnType, sqlSourceType, "getTxn", getStatusValue(err)).Inc()
			return res, err
		}
	}
	metrics.TxnDBCounter.WithLabelValues(txnType, kvSourceType, "getTxn", getStatusValue(err)).Inc()
	return r, err
}

func (bb *TxDB) GetTxns(txnHashes []Hash) (stxns []*SignedTxn, err error) {
	if bb.nodeType == LightNode {
		return nil, nil
	}
	txns := make([]*SignedTxn, 0)
	for _, txnHash := range txnHashes {
		result, err := bb.GetTxn(txnHash)
		if err != nil {
			return nil, err
		}
		if result == nil {
			continue
		}
		txns = append(txns, result)
	}
	return txns, nil
}

func (bb *TxDB) ExistTxn(txnHash Hash) bool {
	if bb.nodeType == LightNode {
		return false
	}
	find := bb.txnKV.ExistTxn(txnHash)
	if !find {
		var records []TxnDBSchema
		err := bb.db.Db().Raw("select value from txndb where type = ? and hash_key = ?", "txn", txnHash.String()).Find(&records).Error
		if err == nil && len(records) > 0 {
			return true
		}
	}
	return find
}

func (bb *TxDB) SetTxns(txns []*SignedTxn) (err error) {
	if bb.nodeType == LightNode {
		return nil
	}
	defer func() {
		metrics.TxnDBCounter.WithLabelValues(txnType, getSourceTypeValue(bb.enableUseSql), "setTxns", getStatusValue(err)).Inc()
	}()
	if bb.enableUseSql {
		for _, txn := range txns {
			txbyt, err := txn.Encode()
			if err != nil {
				logrus.Errorf("TxDB.SetTxns set tx(%s) failed: %v", txn.TxnHash.String(), err)
				return err
			}
			if err := bb.db.Db().Exec("insert into txndb (type, hash_key, value) values (?,?,?)", "txn", txn.TxnHash.String(), txbyt).Error; err != nil {
				logrus.Errorf("Insert TxDB.SetTxns tx(%s) failed: %v", txn.TxnHash.String(), err)
				return err
			}
		}
		return nil
	}
	return bb.txnKV.SetTxns(txns)
}

func (bb *TxDB) SetReceipts(receipts map[Hash]*Receipt) (err error) {
	if bb.enableUseSql {
		for txHash, receipt := range receipts {
			if err := bb.SetReceipt(txHash, receipt); err != nil {
				return err
			}
		}
		return nil
	}
	return bb.receiptKV.SetReceipts(receipts)
}

func (bb *TxDB) SetReceipt(txHash Hash, receipt *Receipt) (err error) {
	defer func() {
		metrics.TxnDBCounter.WithLabelValues(receiptType, getSourceTypeValue(bb.enableUseSql), "setReceipt", getStatusValue(err)).Inc()
	}()
	if bb.enableUseSql {
		byt, err := receipt.Encode()
		if err != nil {
			return err
		}
		if err := bb.db.Db().Exec("insert into txndb (type, hash_key, value) values (?,?,?)", "receipt", txHash.String(), byt).Error; err != nil {
			return err
		}
		return nil
	}
	return bb.receiptKV.SetReceipt(txHash, receipt)
}

func (bb *TxDB) GetReceipt(txHash Hash) (rec *Receipt, err error) {
	r, err := bb.receiptKV.GetReceipt(txHash)
	if err == nil && r == nil {
		var records []TxnDBSchema
		err := bb.db.Db().Raw("select value from txndb where type = ? and hash_key = ?", "receipt", txHash.String()).Find(&records).Error
		if err == nil && len(records) > 0 {
			receipt := new(Receipt)
			err = receipt.Decode(records[0].Value)
			metrics.TxnDBCounter.WithLabelValues(receiptType, sqlSourceType, "getReceipt", getStatusValue(err)).Inc()
			return receipt, err
		}
	}
	metrics.TxnDBCounter.WithLabelValues(receiptType, kvSourceType, "getReceipt", getStatusValue(err)).Inc()
	return r, err
}

type receipttxnkvdb struct {
	sync.RWMutex
	receiptKV kv.KV
}

func (r *receipttxnkvdb) GetReceipt(txHash Hash) (*Receipt, error) {
	r.RLock()
	defer r.RUnlock()
	byt, err := r.receiptKV.Get(txHash.Bytes())
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

func (r *receipttxnkvdb) SetReceipt(txHash Hash, receipt *Receipt) error {
	r.Lock()
	defer r.Unlock()
	byt, err := receipt.Encode()
	if err != nil {
		return err
	}
	return r.receiptKV.Set(txHash.Bytes(), byt)
}

func (r *receipttxnkvdb) SetReceipts(receipts map[Hash]*Receipt) error {
	r.Lock()
	defer r.Unlock()
	kvtx, err := r.receiptKV.NewKvTxn()
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
