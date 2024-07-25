package txpool

import (
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/txdb"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
	"testing"
)

func initTxpool(t *testing.T) *TxPool {
	cfg := config.InitDefaultCfg()
	kvdb, err := kv.NewKvdb(&config.KVconf{
		KvType: "bolt",
		Path:   "./test-txpool.db",
		Hosts:  nil,
	})
	if err != nil {
		t.Fatal("init kvdb error: ", err)
	}
	base := txdb.NewTxDB(common.FullNode, kvdb)
	return WithDefaultChecks(common.FullNode, &cfg.Txpool, base)
}

func TestCheckPoolSize(t *testing.T) {
	pool := initTxpool(t)
	pool.poolSize = 1
	err := pool.Insert(tx1)
	if err != nil {
		t.Fatalf("Insert tx1 failed: %v", err)
	}
	err = pool.Insert(tx2)
	if err != nil {
		assert.Equal(t, yerror.PoolOverflow, err, err)
	}
}

func TestCheckTxnSize(t *testing.T) {
	pool := initTxpool(t)
	pool.TxnMaxSize = 1
	err := pool.Insert(tx1)
	if err != nil {
		assert.Equal(t, yerror.TxnTooLarge, err, err)
	}
}

func TestPackFor(t *testing.T) {
	pool := initTxpool(t)
	err := pool.Insert(tx1)
	if err != nil {
		t.Fatalf("Insert tx1 failed: %v", err)
	}
	err = pool.Insert(tx2)
	if err != nil {
		t.Fatalf("Insert tx2 failed: %v", err)
	}
	err = pool.Insert(tx3)
	if err != nil {
		t.Fatalf("Insert tx3 failed: %v", err)
	}

	txns, err := pool.PackFor(3, func(tx *types.SignedTxn) bool {
		return *tx.GetCaller() == caller2
	})
	if err != nil {
		t.Fatalf("pack txns failed: %v", err)
	}
	assert.Equal(t, txns[0], tx3)
}
