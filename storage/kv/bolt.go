package kv

import (
	"bytes"
	"go.etcd.io/bbolt"
)

type boltKV struct {
	db *bbolt.DB
}

var bucket = []byte("yu")

func NewBolt(fpath string) (*boltKV, error) {
	db, err := bbolt.Open(fpath, 0666, nil)
	if err != nil {
		return nil, err
	}
	return &boltKV{db: db}, nil
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

func (b *boltKV) Iter(keyPrefix []byte) (Iterator, error) {
	var c *bbolt.Cursor
	err := b.db.View(func(tx *bbolt.Tx) error {
		c = tx.Bucket(bucket).Cursor()
		c.Seek(keyPrefix)
		return nil
	})
	return &boltIterator{
		c: c,
		keyPrefix: keyPrefix,
	}, err
}

type boltIterator struct {
	keyPrefix []byte
	key []byte
	value []byte
	c *bbolt.Cursor
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
