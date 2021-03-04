package kv

import (
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
	goctx "golang.org/x/net/context"
	"yu/storage"
)

type tiKV struct {
	store kv.Storage
}

func NewTiKV(path string) (*tiKV, error) {
	driver := tikv.Driver{}
	store, err := driver.Open(path)
	if err != nil {
		return nil, err
	}
	return &tiKV{
		store: store,
	}, nil
}

func (*tiKV) Type() storage.StoreType {
	return storage.Server
}

func (*tiKV) Kind() storage.StoreKind {
	return storage.KV
}

func (t *tiKV) Get(key []byte) ([]byte, error) {
	tx, err := t.store.Begin()
	if err != nil {
		return nil, err
	}
	return tx.Get(goctx.Background(), key)
}

func (t *tiKV) Set(key, value []byte) error {
	tx, err := t.store.Begin()
	if err != nil {
		return err
	}
	err = tx.Set(key, value)
	if err != nil {
		return err
	}
	return tx.Commit(goctx.Background())
}

func (t *tiKV) Exist(key []byte) bool {
	value, _ := t.Get(key)
	return value != nil
}

func (t *tiKV) Iter(key []byte) (Iterator, error) {
	tx, err := t.store.Begin()
	if err != nil {
		return nil, err
	}
	iter, err := tx.Iter(key, nil)
	if err != nil {
		return nil, err
	}
	return &tikvIterator{
		iter: iter,
	}, nil
}

func (t *tiKV) NewKvTxn() (KvTxn, error) {
	tx, err := t.store.Begin()
	if err != nil {
		return nil, err
	}
	return &tikvTxn{tx: tx}, nil
}

type tikvIterator struct {
	iter kv.Iterator
}

func (ti *tikvIterator) Valid() bool {
	return ti.iter.Valid()
}

func (ti *tikvIterator) Next() error {
	return ti.iter.Next()
}

func (ti *tikvIterator) Entry() ([]byte, []byte, error) {
	return ti.iter.Key()[:], ti.iter.Value(), nil
}

func (ti *tikvIterator) Close() {
	ti.iter.Close()
}

type tikvTxn struct {
	tx kv.Transaction
}

func (tt *tikvTxn) Get(key []byte) ([]byte, error) {
	return tt.tx.Get(goctx.Background(), key)
}

func (tt *tikvTxn) Set(key, value []byte) error {
	return tt.tx.Set(key, value)
}

func (tt *tikvTxn) Commit() error {
	return tt.tx.Commit(goctx.Background())
}

func (tt *tikvTxn) Rollback() error {
	return tt.tx.Rollback()
}
