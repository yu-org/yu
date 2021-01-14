package mpt

import (
	"bytes"
	"testing"
	"yu/storage/kv"
)

func TestDataBaseSetPutandGet(t *testing.T) {
	cfg := &kv.KVconf{
		KVtype: "badger",
		Path:   "./testdb",
	}
	db, err := NewNodeBase(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	var expGet = []byte("value")
	db.Put([]byte("key"), expGet)
	var toGet []byte
	toGet, err = db.Get([]byte("key"))
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(expGet, toGet) {
		t.Error("Put value is not equal to Getting value..., expecting", expGet, "but,", toGet)
		return
	}
}
