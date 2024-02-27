package txdb

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/receipt"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
)

const (
	Txns    = "txns"
	Results = "results"
)

type TxDB struct {
	nodeType  int
	txnKV     kv.KV
	receiptKV kv.KV
}

func NewTxDB(nodeTyp int, kvdb kv.Kvdb) ItxDB {
	return &TxDB{
		nodeType:  nodeTyp,
		txnKV:     kvdb.New(Txns),
		receiptKV: kvdb.New(Results),
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

func (bb *TxDB) SetReceipts(receipts []*Receipt) error {
	kvtx, err := bb.receiptKV.NewKvTxn()
	if err != nil {
		return err
	}

	for _, receipt := range receipts {
		byt, err := receipt.Encode()
		if err != nil {
			return err
		}
		hash, err := receipt.Hash()
		if err != nil {
			return err
		}
		err = kvtx.Set(hash, byt)
		if err != nil {
			return err
		}
	}

	return kvtx.Commit()
}

func (bb *TxDB) SetReceipt(receipt *Receipt) error {
	byt, err := receipt.Encode()
	if err != nil {
		return err
	}
	hash, err := receipt.Hash()
	if err != nil {
		return err
	}
	return bb.receiptKV.Set(hash, byt)
}
