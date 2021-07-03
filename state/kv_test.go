package state

import (
	"github.com/Lawliet-Chan/yu/config"
	"testing"
)

var TestStateKvCfg = &config.StateKvConf{
	IndexDB: config.KVconf{KvType: "bolt", Path: "./test_state_index.db", Hosts: nil},
	NodeBase: config.KVconf{
		KvType: "bolt",
		Path:   "./test_state_nodebase.db",
		Hosts:  nil,
	},
}

type TestTripod struct{}

func (tt *TestTripod) Name() string {
	return "test-tripod"
}

func TestKvCommit(t *testing.T) {
	statekv, err := NewStateKV(TestStateKvCfg)
	if err != nil {
		panic("new state-kv error: " + err.Error())
	}

	tri := &TestTripod{}

	statekv.Set(tri, []byte("dayu-key"), []byte("dayu-value"))

	statekv.NextTxn()

	stateRoot, err := statekv.Commit()
	if err != nil {
		t.Fatalf("commit state-kv error: %s", err.Error())
	}

	t.Logf("state-root is %s", stateRoot.String())

	value, err := statekv.Get(tri, []byte("dayu-key"))
	if err != nil {
		t.Fatalf("get state-kv error: %s", err.Error())
	}
	t.Logf("value is %s", string(value))
}
