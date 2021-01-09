package kv

import "github.com/dgraph-io/badger"

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

func (bg *badgerKV) Get(key []byte) ([]byte, error) {
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
	return valCopy, err
}

func (bg *badgerKV) Set(key, value []byte) error {
	return bg.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (bg *badgerKV) Iter(key []byte) (Iterator, error) {
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
