package kv

import (
	. "yu/config"
	"yu/storage"
	. "yu/yerror"
)

type KV interface {
	storage.StorageType
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Delete(key []byte) error
	Exist(key []byte) bool
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
	Entry() (key, value []byte, err error)
	Close()
}

type KvTxn interface {
	Get([]byte) ([]byte, error)
	Set(key, value []byte) error
	Delete(key []byte) error
	Commit() error
	Rollback() error
}
