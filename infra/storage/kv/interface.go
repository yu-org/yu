package kv

import (
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage"
)

// todo: We need DB connection pool

type Kvdb interface {
	storage.StorageType
	NewKVInstance(prefix string) KV
	Get(prefix string, key []byte) ([]byte, error)
	Set(prefix string, key []byte, value []byte) error
	Delete(prefix string, key []byte) error
	Exist(prefix string, key []byte) bool
	Iter(prefix string, key []byte) (Iterator, error)
	NewKvTxn(prefix string) (KvTxn, error)
}

func NewKvdb(cfg *KVconf) (Kvdb, error) {
	switch cfg.KvType {
	case "badger":
		return NewBadger(cfg.Path)
	case "bolt":
		return NewBolt(cfg.Path)
	//case "tikv":
	//	return NewTiKV(cfg.Path)

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
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	Delete(key []byte) error
	Commit() error
	Rollback() error
}

func makeKey(prefix string, key []byte) []byte {
	return append([]byte(prefix), key...)
}
