package txdb

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/result"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
)

type LightTxDB struct {
	resultKV kv.KV
}

func (l *LightTxDB) GetTxn(txnHash Hash) (*SignedTxn, error) {
	return nil, nil
}

func (l *LightTxDB) ExistTxn(txnHash Hash) bool {
	return false
}

func (l *LightTxDB) SetTxns(txns []*SignedTxn) error {
	return nil
}

func (l *LightTxDB) SetResults(results []Result) error {
	kvtx, err := l.resultKV.NewKvTxn()
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

func (l *LightTxDB) SetResult(result Result) error {
	byt, err := result.Encode()
	if err != nil {
		return err
	}
	hash, err := result.Hash()
	if err != nil {
		return err
	}
	return l.resultKV.Set(hash.Bytes(), byt)
}
