package txdb

import (
	"github.com/sirupsen/logrus"

	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

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
