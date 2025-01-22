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
	nodeType int

	txnKV     *txnkvdb
	receiptKV *receipttxnkvdb
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

// func (t *txnkvdb) GetTxn(txnHash Hash) (txn *SignedTxn, err error) {
// 	t.RLock()
// 	defer t.RUnlock()
// 	byt, err := t.txnKV.Get(txnHash.Bytes())
// 	if err != nil {
// 		logrus.Errorf("TxDB.GetTxn(%s), t.txnKV.Get(txnHash.Bytes()) failed: %v", txnHash.String(), err)
// 		return nil, err
// 	}
// 	if byt == nil {
// 		return nil, nil
// 	}
// 	return DecodeSignedTxn(byt)
// }

func (t *txnkvdb) GetTxn(txnHash Hash) (txn *SignedTxn, err error) {
	var byt []byte

	for i := 0; i < maxRetries; i++ {
		t.RLock()
		byt, err = t.txnKV.Get(txnHash.Bytes())
		t.RUnlock()
		if err != nil {
			logrus.Debugf("TxDB.GetTxn(%s), t.txnKV.Get(txnHash.Bytes()) failed: %v", txnHash.String(), err)
			return nil, err
		}
		if byt == nil {
			return nil, nil
		}
		txn, err = DecodeSignedTxn(byt)
		if err == nil {
			if i > 0 {
				logrus.Debugf("TxDB.GetTxn(%s), retry %d times, data: %s", txnHash.String(), i, string(byt))
			}
			return txn, nil
		} else {
			logrus.Debugf("TxDB.GetTxn(%s), DecodeSignedTxn failed, data: %s, retry %d times, error: %v", txnHash.String(), string(byt), i, err)
		}
	}

	return nil, err
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
			logrus.Debugf("TxDB.SetTxns set tx(%s) failed: %v", txn.TxnHash.String(), err)
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
	return txdb, nil
}

func (bb *TxDB) GetTxn(txnHash Hash) (stxn *SignedTxn, err error) {
	if bb.nodeType == LightNode {
		return nil, nil
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
	find := bb.txnKV.ExistTxn(txnHash)
	return find
}

func (bb *TxDB) SetTxns(txns []*SignedTxn) (err error) {
	if bb.nodeType == LightNode {
		return nil
	}
	defer func() {
		metrics.TxnDBCounter.WithLabelValues(txnType, getSourceTypeValue(false), "setTxns", getStatusValue(err)).Inc()
	}()
	return bb.txnKV.SetTxns(txns)
}

func (bb *TxDB) SetReceipts(receipts map[Hash]*Receipt) (err error) {
	defer func() {
		metrics.TxnDBCounter.WithLabelValues(receiptType, getSourceTypeValue(false), "setReceipts", getStatusValue(err)).Inc()
	}()
	return bb.receiptKV.SetReceipts(receipts)
}

func (bb *TxDB) SetReceipt(txHash Hash, receipt *Receipt) (err error) {
	defer func() {
		metrics.TxnDBCounter.WithLabelValues(receiptType, getSourceTypeValue(false), "setReceipt", getStatusValue(err)).Inc()
	}()
	return bb.receiptKV.SetReceipt(txHash, receipt)
}

func (bb *TxDB) GetReceipt(txHash Hash) (rec *Receipt, err error) {
	r, err := bb.receiptKV.GetReceipt(txHash)
	metrics.TxnDBCounter.WithLabelValues(receiptType, kvSourceType, "getReceipt", getStatusValue(err)).Inc()
	return r, err
}

type receipttxnkvdb struct {
	sync.RWMutex
	receiptKV kv.KV
}

// func (r *receipttxnkvdb) GetReceipt(txHash Hash) (*Receipt, error) {
// 	r.RLock()
// 	defer r.RUnlock()
// 	byt, err := r.receiptKV.Get(txHash.Bytes())
// 	if err != nil {
// 		logrus.Errorf("TxDB.GetReceipt(%s), failed: %s, error: %v", txHash.String(), string(byt), err)
// 		return nil, err
// 	}
// 	if byt == nil {
// 		return nil, nil
// 	}
// 	receipt := new(Receipt)
// 	err = receipt.Decode(byt)
// 	if err != nil {
// 		logrus.Errorf("TxDB.GetReceipt(%s), Decode failed: %s, error: %v", txHash.String(), string(byt), err)
// 	}
// 	return receipt, err
// }

func (r *receipttxnkvdb) GetReceipt(txHash Hash) (*Receipt, error) {
	var byt []byte
	var err error

	for i := 0; i < maxRetries; i++ {
		r.RLock()
		byt, err = r.receiptKV.Get(txHash.Bytes())
		r.RUnlock()
		if err != nil {
			logrus.Debugf("TxDB.GetReceipt(%s), failed: %s, error: %v", txHash.String(), string(byt), err)
			return nil, err
		}
		if byt == nil {
			return nil, nil
		}
		receipt := new(Receipt)
		err = receipt.Decode(byt)
		if err == nil {
			if i > 0 {
				logrus.Debugf("TxDB.GetReceipt(%s), succeeded after %d retries, data: %s", txHash.String(), i, string(byt))
			}
			return receipt, nil
		} else {
			logrus.Debugf("TxDB.GetReceipt(%s), Decode failed: %s, retry %d times, error: %v", txHash.String(), string(byt), i, err)
		}
	}

	return nil, err
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
