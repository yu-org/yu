package mpt

import (
	"bytes"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage/kv"
	"testing"
)

func TestTrieSetPutandGet(t *testing.T) {
	cfg := &config.KVconf{
		KvType: "bolt",
		Path:   "./testdb",
	}
	kvdb, err := kv.NewKvdb(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	db := NewNodeBase(kvdb)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	var tr *Trie
	tr, err = NewTrie(HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"), db)
	if err != nil {
		t.Error(err)
		return
	}

	var expGet = []byte("value")
	tr.Update([]byte("key"), expGet)
	tr.Update([]byte("kez"), []byte("error"))
	tr.Update([]byte("keyyy"), []byte("error"))
	tr.Update([]byte("keyyyyy"), []byte("error"))
	tr.Update([]byte("ke"), []byte("error"))

	var toGet []byte
	toGet = tr.Get([]byte("key"))
	if !bytes.Equal(expGet, toGet) {
		t.Error("Put value is not equal to Getting value from memory..., expecting", expGet, "but,", toGet)
		return
	}

	var trHash Hash
	trHash, err = tr.Commit(nil)

	tr = nil
	tr, err = NewTrie(trHash, db)
	if err != nil {
		t.Error(err)
		return
	}

	toGet = tr.Get([]byte("key"))
	if !bytes.Equal(expGet, toGet) {
		t.Error("Put value is not equal to Getting value from db..., expecting", expGet, "but,", toGet)
		return
	}
}
