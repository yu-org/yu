package kv

import "github.com/cockroachdb/pebble"

type pebbleKV struct {
	db *pebble.DB
}

func NewPebble(fpath string) (*pebbleKV, error) {
	db, err := pebble.Open(fpath, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &pebbleKV{db: db}, nil
}

func (p *pebbleKV) Get(key []byte) ([]byte, error) {
	value, closer, err := p.db.Get(key)
	if err != nil {
		return nil, err
	}
	if err := closer.Close(); err != nil {
		return nil, err
	}
	return value, nil
}

func (p *pebbleKV) Set(key []byte, value []byte) error {
	return p.db.Set(key, value, pebble.Sync)
}
