package kv

import (
	"github.com/cockroachdb/pebble"
	"github.com/yu-org/yu/infra/storage"
)

type Pebble struct {
	db *pebble.DB
}

func NewPebble(fpath string) (Kvdb, error) {
	db, err := pebble.Open(fpath, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &Pebble{db: db}, nil
}

func (p *Pebble) Type() storage.StoreType {
	return storage.Embedded
}

func (p *Pebble) Kind() storage.StoreKind {
	return storage.KV
}

func (p *Pebble) New(prefix string) KV {
	return NewKV(prefix, p)
}

func (p *Pebble) Get(prefix string, key []byte) ([]byte, error) {
	key = makeKey(prefix, key)
	// pebble only returns ErrNotFound, if no value, we should return nil []byte.
	value, closer, err := p.db.Get(key)
	if err != nil {
		return value, nil
	}
	return value, closer.Close()
}

func (p *Pebble) Set(prefix string, key []byte, value []byte) error {
	key = makeKey(prefix, key)
	return p.db.Set(key, value, pebble.Sync)
}

func (p *Pebble) Delete(prefix string, key []byte) error {
	key = makeKey(prefix, key)
	return p.db.Delete(key, pebble.Sync)
}

func (p *Pebble) Exist(prefix string, key []byte) bool {
	value, _ := p.Get(prefix, key)
	return value != nil
}

func (p *Pebble) Iter(prefix string, key []byte) (Iterator, error) {
	key = makeKey(prefix, key)
	iter, err := p.db.NewIter(&pebble.IterOptions{})
	if err != nil {
		return nil, err
	}
	iter.SeekPrefixGE(key)
	return &PebbleIter{iter: iter}, nil
}

func (p *Pebble) NewKvTxn(prefix string) (KvTxn, error) {
	return &PebbleTxn{
		prefix: prefix,
		batch:  p.db.NewBatch(),
	}, nil
}

type PebbleIter struct {
	iter *pebble.Iterator
}

func (p *PebbleIter) Valid() bool {
	return p.iter.Valid()
}

func (p *PebbleIter) Next() (err error) {
	p.iter.Next()
	return
}

func (p *PebbleIter) Entry() (key, value []byte, err error) {
	key = p.iter.Key()
	value, err = p.iter.ValueAndErr()
	return
}

func (p *PebbleIter) Close() {
	p.iter.Close()
}

type PebbleTxn struct {
	prefix string
	batch  *pebble.Batch
}

func (p *PebbleTxn) Get(key []byte) ([]byte, error) {
	key = makeKey(p.prefix, key)
	value, closer, err := p.batch.Get(key)
	if err != nil {
		return value, nil
	}
	return value, closer.Close()
}

func (p *PebbleTxn) Set(key, value []byte) error {
	key = makeKey(p.prefix, key)
	return p.batch.Set(key, value, &pebble.WriteOptions{Sync: false})
}

func (p *PebbleTxn) Delete(key []byte) error {
	key = makeKey(p.prefix, key)
	return p.batch.Delete(key, &pebble.WriteOptions{Sync: true})
}

func (p *PebbleTxn) Commit() error {
	return p.batch.Commit(&pebble.WriteOptions{Sync: true})
}

func (p *PebbleTxn) Rollback() error {
	p.batch.Reset()
	return nil
}
