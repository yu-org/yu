package tripod_store

import (
	. "yu/storage/kv"
	. "yu/yerror"
)

type TripodKV struct {
	KVs     map[string]KV
	stashes map[string]*KvStash

	defaultKV    KV
	defaultStash *KvStash
}

func (tkv *TripodKV) GetKV(tag string) KV {
	return tkv.KVs[tag]
}

func (tkv *TripodKV) SetKV(tag string, kv KV) {
	tkv.KVs[tag] = kv
}

func (tkv *TripodKV) GetDefaultKV() KV {
	return tkv.defaultKV
}

func (tkv *TripodKV) SetDefaultKV(kv KV) {
	tkv.defaultKV = kv
}

func (tkv *TripodKV) DfGet(key []byte) ([]byte, error) {
	return tkv.defaultKV.Get(key)
}

func (tkv *TripodKV) DfSet(key, value []byte) error {
	err := tkv.stashOldDfState(key)
	if err != nil {
		return err
	}
	return tkv.defaultKV.Set(key, value)
}

func (tkv *TripodKV) DfDelete(key []byte) error {
	err := tkv.stashOldDfState(key)
	if err != nil {
		return err
	}
	return tkv.defaultKV.Delete(key)
}

func (tkv *TripodKV) Get(tag string, key []byte) ([]byte, error) {
	kv, ok := tkv.KVs[tag]
	if !ok {
		return nil, NoTripodKV
	}
	return kv.Get(key)
}

func (tkv *TripodKV) Set(tag string, key, value []byte) error {
	kv, ok := tkv.KVs[tag]
	if !ok {
		return NoTripodKV
	}
	err := tkv.stashOldState(tag, key)
	if err != nil {
		return err
	}
	return kv.Set(key, value)
}

func (tkv *TripodKV) Delete(tag string, key []byte) error {
	kv, ok := tkv.KVs[tag]
	if !ok {
		return NoTripodKV
	}
	err := tkv.stashOldState(tag, key)
	if err != nil {
		return err
	}
	return kv.Delete(key)
}

func (tkv *TripodKV) stashOldDfState(key []byte) error {
	value, err := tkv.defaultKV.Get(key)
	if err != nil {
		return err
	}
	tkv.defaultStash = &KvStash{
		Key:   key,
		Value: value,
	}
	return nil
}

func (tkv *TripodKV) stashOldState(tag string, key []byte) error {
	value, err := tkv.KVs[tag].Get(key)
	if err != nil {
		return err
	}
	tkv.stashes[tag] = &KvStash{
		Key:   key,
		Value: value,
	}
	return nil
}

func (tkv *TripodKV) Commit() {

}

func (tkv *TripodKV) Rollback() {

}

func (tkv *TripodKV) Flush() {
	tkv.defaultStash = nil
	tkv.stashes = nil
}

type KvStash struct {
	Key   []byte
	Value []byte
}
