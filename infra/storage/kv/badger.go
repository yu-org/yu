package kv

import (
	"github.com/dgraph-io/badger"
	"github.com/yu-org/yu/infra/storage"
)

type badgerKV struct {
	db *badger.DB
}

func NewBadger(path string) (*badgerKV, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return &badgerKV{
		db: db,
	}, nil
}

func (b *badgerKV) New(prefix string) KV {
	return NewKV(prefix, b)
}

func (*badgerKV) Type() storage.StoreType {
	return storage.Embedded
}

func (*badgerKV) Kind() storage.StoreKind {
	return storage.KV
}

func (bg *badgerKV) Get(prefix string, key []byte) ([]byte, error) {
	key = makeKey(prefix, key)
	var valCopy []byte
	err := bg.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			valCopy = append(valCopy, val...)
			return nil
		})
		return nil
	})
	if err != nil && err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return valCopy, err
}

func (bg *badgerKV) Set(prefix string, key, value []byte) error {
	key = makeKey(prefix, key)
	return bg.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (bg *badgerKV) Delete(prefix string, key []byte) error {
	key = makeKey(prefix, key)
	return bg.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (bg *badgerKV) Exist(prefix string, key []byte) bool {
	value, _ := bg.Get(prefix, key)
	return value != nil
}

func (bg *badgerKV) Iter(prefix string, key []byte) (Iterator, error) {
	key = makeKey(prefix, key)
	var iter *badger.Iterator
	err := bg.db.View(func(txn *badger.Txn) error {
		iter = txn.NewIterator(badger.DefaultIteratorOptions)
		iter.Seek(key)
		return nil
	})
	return &badgerIterator{
		key:  key,
		iter: iter,
	}, err
}

func (bg *badgerKV) NewKvTxn(prefix string) (KvTxn, error) {
	tx := bg.db.NewTransaction(true)
	return &badgerTxn{
		prefix: prefix,
		tx:     tx,
	}, nil
}

type badgerIterator struct {
	key  []byte
	iter *badger.Iterator
}

func (bgi *badgerIterator) Valid() bool {
	return bgi.iter.ValidForPrefix(bgi.key)
}

func (bgi *badgerIterator) Next() error {
	bgi.iter.Next()
	return nil
}

func (bgi *badgerIterator) Entry() ([]byte, []byte, error) {
	var value []byte
	item := bgi.iter.Item()
	key := item.Key()
	err := item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})
	return key, value, err
}

func (bgi *badgerIterator) Close() {
	bgi.iter.Close()
}

type badgerTxn struct {
	prefix string
	tx     *badger.Txn
}

func (bt *badgerTxn) Get(key []byte) ([]byte, error) {
	key = makeKey(bt.prefix, key)
	item, err := bt.tx.Get(key)
	if err != nil {
		return nil, err
	}
	var value []byte
	err = item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})
	return value, err
}

func (bt *badgerTxn) Set(key, value []byte) error {
	key = makeKey(bt.prefix, key)
	return bt.tx.Set(key, value)
}

func (bt *badgerTxn) Delete(key []byte) error {
	key = makeKey(bt.prefix, key)
	return bt.tx.Delete(key)
}

func (bt *badgerTxn) Commit() error {
	return bt.tx.Commit()
}

func (bt *badgerTxn) Rollback() error {
	bt.tx.Discard()
	return nil
}
