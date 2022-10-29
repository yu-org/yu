package txdb

import (
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
	txnKV    kv.KV
	resultKV kv.KV
}

func NewTxDB(nodeTyp NodeType, kvdb kv.Kvdb) ItxDB {
	if nodeTyp == LightNode {
		return &LightTxDB{resultKV: kvdb.New(Results)}
	}
	return &TxDB{
		txnKV:    kvdb.New(Txns),
		resultKV: kvdb.New(Results),
	}
}

func (bb *TxDB) GetTxn(txnHash Hash) (*SignedTxn, error) {
	byt, err := bb.txnKV.Get(txnHash.Bytes())
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxn(byt)
}

func (bb *TxDB) ExistTxn(txnHash Hash) bool {
	return bb.txnKV.Exist(txnHash.Bytes())
}

func (bb *TxDB) SetTxns(txns []*SignedTxn) error {
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
	kvtx, err := bb.resultKV.NewKvTxn()
	if err != nil {
		return err
	}
	for _, result := range results {
		byt, err := result.Encode()
		if err != nil {
			return err
		}
		hash, err := result.Hash()
		if err != nil {
			return err
		}
		err = kvtx.Set(hash.Bytes(), byt)
	}
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
