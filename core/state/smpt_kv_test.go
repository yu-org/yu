package state

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
	"path/filepath"
	"testing"
)

var kvcfg = &config.KVconf{
	KvType: "bolt",
	Path:   "./test-smpt-kv.db",
	Hosts:  nil,
}

var (
	key1   = []byte("dayu-key1")
	value1 = []byte("dayu-value1")
	key2   = []byte("dayu-key2")
	value2 = []byte("dayu-value2")
)

type TestTripod1 struct{}

// Name identifies the first test namespace.
func (tt *TestTripod1) Name() string {
	return "1"
}

type TestTripod2 struct{}

// Name identifies the second test namespace.
func (TestTripod2) Name() string {
	return "2"
}

// TestKvCommit checks the current KV lifecycle using its block-based API.
func TestKvCommit(t *testing.T) {
	cfg := *kvcfg
	cfg.Path = filepath.Join(t.TempDir(), "state.db")
	kvdb, err := kv.NewKvdb(&cfg)
	assert.NoError(t, err)
	statekv := NewSpmtKV(nil, kvdb)
	block := &types.Block{Header: &types.Header{Hash: NullHash}}
	statekv.StartBlock(block)

	tri1 := new(TestTripod1)
	tri2 := new(TestTripod2)

	statekv.Set(tri1, key1, value1)
	statekv.Set(tri2, key2, value2)

	statekv.NextTxn()

	_, err = statekv.Commit()
	if err != nil {
		t.Fatalf("apply state-kv error: %s", err.Error())
	}

	statekv.FinalizeBlock(block)

	value, err := statekv.Get(tri1, key1)
	assert.NoError(t, err, "get key1 state-kv error")
	assert.Equal(t, value1, value)

	value, err = statekv.Get(tri2, key2)
	assert.NoError(t, err, "get key2 state-kv error")
	assert.Equal(t, value2, value)

	value, err = statekv.GetByBlockHash(tri1, key1, block)
	assert.NoError(t, err, "get key1 state-kv by blockHash error")
	assert.Equal(t, value1, value)

	value, err = statekv.GetByBlockHash(tri2, key2, block)
	assert.NoError(t, err, "get key2 state-kv by blockHash error")
	assert.Equal(t, value2, value)

}
