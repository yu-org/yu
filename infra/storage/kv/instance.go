package kv

type KvInstance struct {
	prefix string
	kvdb   Kvdb
}

func NewKV(prefix string, kvdb Kvdb) KV {
	return &KvInstance{
		prefix: prefix,
		kvdb:   kvdb,
	}
}

func (k *KvInstance) Get(key []byte) ([]byte, error) {
	return k.kvdb.Get(k.prefix, key)
}

func (k *KvInstance) Set(key []byte, value []byte) error {
	return k.kvdb.Set(k.prefix, key, value)
}

func (k *KvInstance) Delete(key []byte) error {
	return k.kvdb.Delete(k.prefix, key)
}

func (k *KvInstance) Exist(key []byte) bool {
	return k.kvdb.Exist(k.prefix, key)
}

func (k *KvInstance) Iter(key []byte) (Iterator, error) {
	return k.kvdb.Iter(k.prefix, key)
}

func (k *KvInstance) NewKvTxn() (KvTxn, error) {
	return k.kvdb.NewKvTxn(k.prefix)
}
