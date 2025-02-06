package txdb

import (
	"sync"

	"github.com/sirupsen/logrus"

	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/metrics"
)

const (
	Txns       = "txns"
	Results    = "results"
	maxRetries = 5
)

type TxDB struct {
	enableSqlite bool
	nodeType     int

	txnKV     *txnkvdb
	receiptKV *receipttxnkvdb

	txnSqlite     *txnSqliteStorage
	receiptSqlite *receiptSqliteStorage
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

func NewTxDB(nodeTyp int, kvdb kv.Kvdb, txnConf config.TxnConf) (ItxDB, error) {
	txdb := &TxDB{
		enableSqlite: txnConf.EnableSqliteStorage,
		nodeType:     nodeTyp,
		txnKV:        &txnkvdb{txnKV: kvdb.New(Txns)},
		receiptKV:    &receipttxnkvdb{receiptKV: kvdb.New(Results)},
	}
	if txdb.enableSqlite {
		txdb.txnSqlite = &txnSqliteStorage{}
		if err := txdb.txnSqlite.initdb(); err != nil {
			return nil, err
		}
		txdb.receiptSqlite = &receiptSqliteStorage{}
		if err := txdb.receiptSqlite.initdb(); err != nil {
			return nil, err
		}
	}
	return txdb, nil
}

func (bb *TxDB) GetTxn(txnHash Hash) (stxn *SignedTxn, err error) {
	if bb.nodeType == LightNode {
		return nil, nil
	}
	if bb.enableSqlite {
		txn, err := bb.txnSqlite.GetTxn(txnHash)
		if err != nil {
			metrics.TxnDBCounter.WithLabelValues(txnType, sqlSourceType, "getTxn", getStatusValue(err)).Inc()
			return nil, err
		}
		if txn != nil {
			metrics.TxnDBCounter.WithLabelValues(txnType, sqlSourceType, "getTxn", getStatusValue(err)).Inc()
			return txn, nil
		}
	}
	r, err := bb.txnKV.GetTxn(txnHash)
	if err != nil {
		logrus.Debugf("TxDB.GetTxn(%s), failed: %v", txnHash.String(), err)
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
	if bb.enableSqlite {
		exists := bb.txnSqlite.ExistTxn(txnHash)
		if exists {
			return exists
		}
	}
	find := bb.txnKV.ExistTxn(txnHash)
	return find
}

func (bb *TxDB) SetTxns(txns []*SignedTxn) (err error) {
	if bb.nodeType == LightNode {
		return nil
	}
	defer func() {
		metrics.TxnDBCounter.WithLabelValues(txnType, getSourceTypeValue(bb.enableSqlite), "setTxns", getStatusValue(err)).Inc()
	}()
	if bb.enableSqlite {
		return bb.txnSqlite.SetTxns(txns)
	}
	return bb.txnKV.SetTxns(txns)
}

func (bb *TxDB) SetReceipts(receipts map[Hash]*Receipt) (err error) {
	defer func() {
		metrics.TxnDBCounter.WithLabelValues(receiptType, getSourceTypeValue(bb.enableSqlite), "setReceipts", getStatusValue(err)).Inc()
	}()
	if bb.enableSqlite {
		return bb.receiptSqlite.SetReceipts(receipts)
	}
	return bb.receiptKV.SetReceipts(receipts)
}

func (bb *TxDB) SetReceipt(txHash Hash, receipt *Receipt) (err error) {
	defer func() {
		metrics.TxnDBCounter.WithLabelValues(receiptType, getSourceTypeValue(bb.enableSqlite), "setReceipt", getStatusValue(err)).Inc()
	}()
	if bb.enableSqlite {
		return bb.receiptSqlite.SetReceipt(txHash, receipt)
	}
	return bb.receiptKV.SetReceipt(txHash, receipt)
}

func (bb *TxDB) GetReceipt(txHash Hash) (rec *Receipt, err error) {
	if bb.enableSqlite {
		r, err := bb.receiptSqlite.GetReceipt(txHash)
		if err != nil {
			metrics.TxnDBCounter.WithLabelValues(receiptType, sqlSourceType, "getReceipt", getStatusValue(err)).Inc()
			return nil, err
		}
		if r != nil {
			metrics.TxnDBCounter.WithLabelValues(receiptType, sqlSourceType, "getReceipt", getStatusValue(err)).Inc()
			return r, nil
		}
	}
	r, err := bb.receiptKV.GetReceipt(txHash)
	metrics.TxnDBCounter.WithLabelValues(receiptType, kvSourceType, "getReceipt", getStatusValue(err)).Inc()
	return r, err
}

type receipttxnkvdb struct {
	sync.RWMutex
	receiptKV kv.KV
}
