package mpt

import (
	"bytes"
	"testing"
)

func TestDataBaseSetPutandGet(t *testing.T) {
	db, err := NewNodeBase("./testdb")
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