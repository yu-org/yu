package mpt

import (
	"bytes"
	"testing"
	"yu/config"
)

func TestDataBaseSetPutandGet(t *testing.T) {
	cfg := &config.KVconf{
		KvType: "badger",
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
