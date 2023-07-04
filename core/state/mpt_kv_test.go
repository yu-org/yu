package state

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage/kv"
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
	return "test-tripod-1"
}

type TestTripod2 struct{}

func (TestTripod2) Name() string {
	return "test-tripod-2"
}

func TestKvCommit(t *testing.T) {
	kvdb, err := kv.NewKvdb(kvcfg)
	statekv := NewMptKV(kvdb)

	tri1 := &TestTripod1{}
	tri2 := &TestTripod2{}

	statekv.Set(tri1, key1, value1)
	statekv.Set(tri2, key2, value2)

	statekv.NextTxn()

	stateRoot, err := statekv.Commit()
	if err != nil {
		t.Fatalf("commit state-kv error: %s", err.Error())
	}

	statekv.FinalizeBlock(NullHash)

	t.Logf("state-root is %s", stateRoot.String())

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
