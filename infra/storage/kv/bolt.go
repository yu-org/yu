package kv

import (
	"bytes"
	"github.com/yu-org/yu/infra/storage"
	"go.etcd.io/bbolt"
)

type boltKV struct {
	db *bbolt.DB
}

var bucket = []byte("github.com/yu-org/yu")

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

func (b *boltKV) New(prefix string) KV {
	return NewKV(prefix, b)
}

func (*boltKV) Type() storage.StoreType {
	return storage.Embedded
}

func (*boltKV) Kind() storage.StoreKind {
	return storage.KV
}

func (b *boltKV) Get(prefix string, key []byte) ([]byte, error) {
	key = makeKey(prefix, key)
	var value []byte
	err := b.db.View(func(tx *bbolt.Tx) error {
		bu := tx.Bucket(bucket)
		value = bu.Get(key)
		return nil
	})
	return value, err
}

func (b *boltKV) Set(prefix string, key []byte, value []byte) error {
	key = makeKey(prefix, key)
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).Put(key, value)
	})
}

func (b *boltKV) Delete(prefix string, key []byte) error {
	key = makeKey(prefix, key)
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).Delete(key)
	})
}

func (b *boltKV) Exist(prefix string, key []byte) bool {
	value, _ := b.Get(prefix, key)
	return value != nil
}

func (b *boltKV) Iter(prefix string, key []byte) (Iterator, error) {
	key = makeKey(prefix, key)
	var c *bbolt.Cursor
	err := b.db.View(func(tx *bbolt.Tx) error {
		c = tx.Bucket(bucket).Cursor()
		c.Seek(key)
		return nil
	})
	return &boltIterator{
		c:         c,
		keyPrefix: key,
	}, err
}

func (b *boltKV) NewKvTxn(prefix string) (KvTxn, error) {
	tx, err := b.db.Begin(true)
	if err != nil {
		return nil, err
	}
	return &boltTxn{
		prefix: prefix,
		tx:     tx,
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
	prefix string
	tx     *bbolt.Tx
}

func (bot *boltTxn) Get(key []byte) ([]byte, error) {
	key = makeKey(bot.prefix, key)
	return bot.tx.Bucket(bucket).Get(key), nil
}

func (bot *boltTxn) Set(key, value []byte) error {
	key = makeKey(bot.prefix, key)
	return bot.tx.Bucket(bucket).Put(key, value)
}

func (bot *boltTxn) Delete(key []byte) error {
	key = makeKey(bot.prefix, key)
	return bot.tx.Bucket(bucket).Delete(key)
}

func (bot *boltTxn) Commit() error {
	return bot.tx.Commit()
}

func (bot *boltTxn) Rollback() error {
	return bot.tx.Rollback()
}
