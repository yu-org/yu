package kv

import (
	"github.com/pkg/errors"
)

var NoKvdbType = errors.New("no kvdb type")

type KV interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Iter(key []byte) (Iterator, error)
}

func NewKV(cfg *KVconf) (KV, error) {
	switch cfg.KVtype {
	case "badger":
		return NewBadger(cfg.Path)
	case "bolt":
		return NewBolt(cfg.Path)
	case "tikv":
		return NewTiKV(cfg.Path)

	default:
		return nil, NoKvdbType
	}
}

type Iterator interface {
	Valid() bool
	Next() error
	Entry() ([]byte, []byte, error)
	Close()
}
