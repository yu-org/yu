package state

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/infra/trie/mpt"
	"os"
	"testing"
)

var kvcfg = &config.KVconf{
	KvType: "bolt",
	Path:   "./test-mpt-kv.db",
	Hosts:  nil,
}

var (
	key1   = []byte("dayu-key1")
	value1 = []byte("dayu-value1")
	key2   = []byte("dayu-key2")
	value2 = []byte("dayu-value2")
)

type TestTripod1 struct{}

func (tt *TestTripod1) Name() string {
	return "1"
}

type TestTripod2 struct{}

func (TestTripod2) Name() string {
	return "2"
}

func TestMPT(t *testing.T) {
	kvdb, err := kv.NewKvdb(kvcfg)
	assert.NoError(t, err)

	nodeBase := mpt.NewNodeBase(kvdb)
	tr, err := mpt.NewTrie(NullHash, nodeBase)
	assert.NoError(t, err)

	assert.NoError(t, tr.TryUpdate(key1, value1))
	assert.NoError(t, tr.TryUpdate(key2, value2))
	stateRoot, err := tr.Commit(nil)
	assert.NoError(t, err)

	tr2, err := mpt.NewTrie(stateRoot, nodeBase)
	assert.NoError(t, err)

	assert.Equal(t, value1, tr2.Get(key1))
	assert.Equal(t, value2, tr2.Get(key2))

	removeTestDB()
}

func TestKvCommit(t *testing.T) {
	kvdb, err := kv.NewKvdb(kvcfg)
	assert.NoError(t, err)
	statekv := NewMptKV(kvdb)

	tri1 := new(TestTripod1)
	tri2 := new(TestTripod2)

	statekv.Set(tri1, key1, value1)
	statekv.Set(tri2, key2, value2)

	statekv.NextTxn()

	_, err = statekv.Commit()
	if err != nil {
		t.Fatalf("commit state-kv error: %s", err.Error())
	}

	statekv.FinalizeBlock(NullHash)

	value, err := statekv.Get(tri1, key1)
	assert.NoError(t, err, "get key1 state-kv error")
	assert.Equal(t, value1, value)

	value, err = statekv.Get(tri2, key2)
	assert.NoError(t, err, "get key2 state-kv error")
	assert.Equal(t, value2, value)

	value, err = statekv.GetByBlockHash(tri1, key1, NullHash)
	assert.NoError(t, err, "get key1 state-kv by blockHash error")
	assert.Equal(t, value1, value)

	value, err = statekv.GetByBlockHash(tri2, key2, NullHash)
	assert.NoError(t, err, "get key2 state-kv by blockHash error")
	assert.Equal(t, value2, value)

	removeTestDB()
}

func removeTestDB() {
	os.RemoveAll(kvcfg.Path)
}
