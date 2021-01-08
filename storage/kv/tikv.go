package kv

import (
	"errors"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
	goctx "golang.org/x/net/context"
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

type tikvIterator struct {
	iter kv.Iterator
}

func (ti *tikvIterator) Next() ([]byte, []byte, error) {
	err := ti.iter.Next()
	if err != nil {
		return nil, nil, err
	}
	if ti.iter.Valid() {
		return ti.iter.Key()[:], ti.iter.Value(), nil
	}
	return nil, nil, EntryInvalid
}

func (ti *tikvIterator) Close() {
	ti.iter.Close()
}
