package kv

import (
	"github.com/pkg/errors"
	. "yu/config"
)

var NoKvdbType = errors.New("no kvdb type")

type KV interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Iter(key []byte) (Iterator, error)
	NewKvTxn() (KvTxn, error)
}

func NewKV(cfg *KVconf) (KV, error) {
	switch cfg.KvType {
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

type KvTxn interface {
	Get([]byte) ([]byte, error)
	Set(key, value []byte) error
	Commit() error
	Rollback() error
}
