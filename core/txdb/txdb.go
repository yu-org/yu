package txdb

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/result"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
)

const (
	Txns    = "txns"
	Results = "results"
)

type TxDB struct {
	nodeType int
	txnKV    kv.KV
	resultKV kv.KV
}

func NewTxDB(nodeTyp int, kvdb kv.Kvdb) ItxDB {
	return &TxDB{
		nodeType: nodeTyp,
		txnKV:    kvdb.New(Txns),
		resultKV: kvdb.New(Results),
	}
}

func (bb *TxDB) GetTxn(txnHash Hash) (*SignedTxn, error) {
	if bb.nodeType == LightNode {
		return nil, nil
	}
	byt, err := bb.txnKV.Get(txnHash.Bytes())
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxn(byt)
}

func (bb *TxDB) ExistTxn(txnHash Hash) bool {
	if bb.nodeType == LightNode {
		return false
	}
	return bb.txnKV.Exist(txnHash.Bytes())
}

func (bb *TxDB) SetTxns(txns []*SignedTxn) error {
	if bb.nodeType == LightNode {
		return nil
	}
	kvtx, err := bb.txnKV.NewKvTxn()
	if err != nil {
		return err
	}
	for _, txn := range txns {
		txbyt, err := txn.Encode()
		if err != nil {
			return err
		}
		err = kvtx.Set(txn.TxnHash.Bytes(), txbyt)
		if err != nil {
			return err
		}
	}
	return kvtx.Commit()
}

func (bb *TxDB) SetResults(results []Result) error {
	if len(results) == 0 {
		return nil
	}
	kvtx, err := bb.resultKV.NewKvTxn()
	if err != nil {
		return err
	}

	var keyLens, valueLens int
	for _, result := range results {
		byt, err := result.Encode()
		if err != nil {
			return err
		}
		hash, err := result.Hash()
		if err != nil {
			return err
		}
		keyLens += len(hash.Bytes())
		valueLens += len(byt)
		err = kvtx.Set(hash.Bytes(), byt)
		if err != nil {
			return err
		}
	}

	logrus.Infof("--- key length is %d, value length is %d ", keyLens, valueLens)
	return kvtx.Commit()
}

func (bb *TxDB) SetResult(result Result) error {
	byt, err := result.Encode()
	if err != nil {
		return err
	}
	hash, err := result.Hash()
	if err != nil {
		return err
	}
	return bb.resultKV.Set(hash.Bytes(), byt)
}
