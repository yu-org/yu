package txdb

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/result"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
)

type TxDB struct {
	kvdb kv.KV
}

const (
	Txns    = "txns"
	Results = "results"
)

func NewTxDB(kvdb kv.KV) *TxDB {
	return &TxDB{
		kvdb: kvdb,
	}
}

func (bb *TxDB) GetTxn(txnHash Hash) (*SignedTxn, error) {
	byt, err := bb.kvdb.Get(Txns, txnHash.Bytes())
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxn(byt)
}

func (bb *TxDB) ExistTxn(txnHash Hash) bool {
	return bb.kvdb.Exist(Txns, txnHash.Bytes())
}

func (bb *TxDB) SetTxns(txns []*SignedTxn) error {
	kvtx, err := bb.kvdb.NewKvTxn()
	if err != nil {
		return err
	}
	for _, txn := range txns {
		txbyt, err := txn.Encode()
		if err != nil {
			return err
		}
		err = kvtx.Set(Txns, txn.TxnHash.Bytes(), txbyt)
		if err != nil {
			return err
		}
	}
	return kvtx.Commit()
}

func (bb *TxDB) SetResults(results []Result) error {
	kvtx, err := bb.kvdb.NewKvTxn()
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
		err = kvtx.Set(Results, hash.Bytes(), byt)
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
	return bb.kvdb.Set(Results, hash.Bytes(), byt)
}
