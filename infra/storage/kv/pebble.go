package kv

import (
	"sync"

	"github.com/cockroachdb/pebble"

	"github.com/yu-org/yu/infra/storage"
)

type Pebble struct {
	sync.Mutex
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
	p.Lock()
	defer p.Unlock()
	key = makeKey(prefix, key)
	// pebble only returns ErrNotFound, if no value, we should return nil []byte.
	value, closer, err := p.db.Get(key)
	if err != nil {
		return value, nil
	}
	return value, closer.Close()
}

func (p *Pebble) Set(prefix string, key []byte, value []byte) error {
	p.Lock()
	defer p.Unlock()
	key = makeKey(prefix, key)
	return p.db.Set(key, value, pebble.Sync)
}

func (p *Pebble) Delete(prefix string, key []byte) error {
	p.Lock()
	defer p.Unlock()
	key = makeKey(prefix, key)
	return p.db.Delete(key, pebble.Sync)
}

func (p *Pebble) Exist(prefix string, key []byte) bool {
	value, _ := p.Get(prefix, key)
	return value != nil
}
