package txdb

import (
	"sync"
	"time"

	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/metrics"
)

const (
	Txns       = "txns"
	Results    = "results"
	maxRetries = 20
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
	sync.Mutex
	txnKV kv.KV
}

const (
	kvSourceType  = "kv"
	txnType       = "txn"
	receiptType   = "receipt"
	successStatus = "success"
	errStatus     = "err"
	pbErrStatus   = "pbErr"
)

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

func (bb *TxDB) getTxn(txnHash Hash) (stxn *SignedTxn, err error) {
	r, err := bb.txnKV.GetTxn(txnHash)
	if err != nil {
		return nil, err
	}
	return r, err
}

func (bb *TxDB) GetTxn(txnHash Hash) (stxn *SignedTxn, err error) {
	if bb.nodeType == LightNode {
		return nil, nil
	}
	start := time.Now()
	defer func() {
		metrics.TxnDBDurationHistogram.WithLabelValues(txnType, kvSourceType, "getTxn", getStatusValue(err)).Observe(float64(time.Since(start).Microseconds()))
	}()
	txn, err := bb.getTxn(txnHash)
	if err != nil {
		metrics.TxnDBCounter.WithLabelValues(txnType, kvSourceType, "getTxn", errStatus).Inc()
		return nil, err
	}
	metrics.TxnDBCounter.WithLabelValues(txnType, kvSourceType, "getTxn", successStatus).Inc()
	return txn, nil
}

func (bb *TxDB) GetTxns(txnHashes []Hash) (stxns []*SignedTxn, err error) {
	if bb.nodeType == LightNode {
		return nil, nil
	}
	start := time.Now()
	defer func() {
		metrics.TxnDBDurationHistogram.WithLabelValues(txnType, kvSourceType, "getTxns", getStatusValue(err)).Observe(float64(time.Since(start).Microseconds()))
	}()
	txns := make([]*SignedTxn, 0)
	for _, txnHash := range txnHashes {
		result, err := bb.getTxn(txnHash)
		if err != nil {
			metrics.TxnDBCounter.WithLabelValues(txnType, kvSourceType, "getTxns", errStatus).Inc()
			return nil, err
		}
		if result == nil {
			continue
		}
		txns = append(txns, result)
	}
	metrics.TxnDBCounter.WithLabelValues(txnType, kvSourceType, "getTxns", successStatus).Inc()
	return txns, nil
}

func (bb *TxDB) ExistTxn(txnHash Hash) bool {
	if bb.nodeType == LightNode {
		return false
	}
	start := time.Now()
	defer func() {
		metrics.TxnDBDurationHistogram.WithLabelValues(txnType, kvSourceType, "exist", successStatus).Observe(float64(time.Since(start).Microseconds()))
	}()
	find := bb.txnKV.ExistTxn(txnHash)
	metrics.TxnDBCounter.WithLabelValues(txnType, kvSourceType, "exist", successStatus).Inc()
	return find
}

func (bb *TxDB) SetTxns(txns []*SignedTxn) (err error) {
	if bb.nodeType == LightNode {
		return nil
	}
	start := time.Now()
	defer func() {
		metrics.TxnDBDurationHistogram.WithLabelValues(txnType, kvSourceType, "setTxns", getStatusValue(err)).Observe(float64(time.Since(start).Microseconds()))
		metrics.TxnDBCounter.WithLabelValues(txnType, kvSourceType, "setTxns", getStatusValue(err)).Inc()
	}()
	return bb.txnKV.SetTxns(txns)
}

func (bb *TxDB) SetReceipts(receipts map[Hash]*Receipt) (err error) {
	start := time.Now()
	defer func() {
		metrics.TxnDBDurationHistogram.WithLabelValues(receiptType, kvSourceType, "setReceipts", getStatusValue(err)).Observe(float64(time.Since(start).Microseconds()))
		metrics.TxnDBCounter.WithLabelValues(receiptType, kvSourceType, "setReceipts", getStatusValue(err)).Inc()
	}()
	return bb.receiptKV.SetReceipts(receipts)
}

func (bb *TxDB) SetReceipt(txHash Hash, receipt *Receipt) (err error) {
	start := time.Now()
	defer func() {
		metrics.TxnDBDurationHistogram.WithLabelValues(receiptType, kvSourceType, "setReceipt", getStatusValue(err)).Observe(float64(time.Since(start).Microseconds()))
		metrics.TxnDBCounter.WithLabelValues(receiptType, kvSourceType, "setReceipt", getStatusValue(err)).Inc()
	}()
	return bb.receiptKV.SetReceipt(txHash, receipt)
}

func (bb *TxDB) GetReceipt(txHash Hash) (rec *Receipt, err error) {
	start := time.Now()
	defer func() {
		metrics.TxnDBDurationHistogram.WithLabelValues(receiptType, kvSourceType, "getReceipt", getStatusValue(err)).Observe(float64(time.Since(start).Microseconds()))
	}()
	r, err := bb.receiptKV.GetReceipt(txHash)
	if err != nil {
		if pbErr, ok := err.(PebbleGetErr); ok {
			metrics.TxnDBCounter.WithLabelValues(receiptType, kvSourceType, "getReceipt", pbErrStatus).Inc()
			return nil, pbErr.err
		}
		metrics.TxnDBCounter.WithLabelValues(receiptType, kvSourceType, "getReceipt", errStatus).Inc()
		return nil, err
	}
	metrics.TxnDBCounter.WithLabelValues(receiptType, kvSourceType, "getReceipt", successStatus).Inc()
	return r, nil
}

type receipttxnkvdb struct {
	sync.Mutex
	receiptKV kv.KV
}
