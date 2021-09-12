package kv

import (
	"bytes"
	"github.com/yu-altar/yu/storage"
	"go.etcd.io/bbolt"
)

type boltKV struct {
	db *bbolt.DB
}

var bucket = []byte("github.com/yu-altar/yu")

func NewBolt(fpath string) (*boltKV, error) {
	db, err := bbolt.Open(fpath, 0666, nil)
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	_, err = tx.CreateBucketIfNotExists(bucket)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &boltKV{db: db}, nil
}

func (*boltKV) Type() storage.StoreType {
	return storage.Embedded
}

func (*boltKV) Kind() storage.StoreKind {
	return storage.KV
}

func (b *boltKV) Get(key []byte) ([]byte, error) {
	var value []byte
	err := b.db.View(func(tx *bbolt.Tx) error {
		bu := tx.Bucket(bucket)
		value = bu.Get(key)
		return nil
	})
	return value, err
}

func (b *boltKV) Set(key []byte, value []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).Put(key, value)
	})
}

func (b *boltKV) Delete(key []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).Delete(key)
	})
}

func (b *boltKV) Exist(key []byte) bool {
	value, _ := b.Get(key)
	return value != nil
}

func (b *boltKV) Iter(keyPrefix []byte) (Iterator, error) {
	var c *bbolt.Cursor
	err := b.db.View(func(tx *bbolt.Tx) error {
		c = tx.Bucket(bucket).Cursor()
		c.Seek(keyPrefix)
		return nil
	})
	return &boltIterator{
		c:         c,
		keyPrefix: keyPrefix,
	}, err
}

func (b *boltKV) NewKvTxn() (KvTxn, error) {
	tx, err := b.db.Begin(true)
	if err != nil {
		return nil, err
	}
	return &boltTxn{
		tx: tx,
	}, nil
}

type boltIterator struct {
	keyPrefix []byte
	key       []byte
	value     []byte
	c         *bbolt.Cursor
}

func (bi *boltIterator) Valid() bool {
	return bi.key != nil && bytes.HasPrefix(bi.key, bi.keyPrefix)
}

func (bi *boltIterator) Next() error {
	bi.key, bi.value = bi.c.Next()
	return nil
}

func (bi *boltIterator) Entry() ([]byte, []byte, error) {
	return bi.key, bi.value, nil
}

func (bi *boltIterator) Close() {

}

type boltTxn struct {
	tx *bbolt.Tx
}

func (bot *boltTxn) Get(key []byte) ([]byte, error) {
	return bot.tx.Bucket(bucket).Get(key), nil
}

func (bot *boltTxn) Set(key, value []byte) error {
	return bot.tx.Bucket(bucket).Put(key, value)
}

func (bot *boltTxn) Delete(key []byte) error {
	return bot.tx.Bucket(bucket).Delete(key)
}

func (bot *boltTxn) Commit() error {
	return bot.tx.Commit()
}

func (bot *boltTxn) Rollback() error {
	return bot.tx.Rollback()
}
