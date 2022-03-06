package txpool

import (
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/core/yudb"
	"testing"
)

func initTxpool() *TxPool {
	cfg := config.InitDefaultCfgWithDir("test-txpool")
	base := yudb.NewYuDB(&cfg.YuDB)
	return WithDefaultChecks(&cfg.Txpool, base)
}

func TestCheckPoolSize(t *testing.T) {
	pool := initTxpool()
	pool.poolSize = 1
	err := pool.Insert(tx1)
	if err != nil {
		t.Fatalf("insert tx1 failed: %v", err)
	}
	err = pool.Insert(tx2)
	if err != nil {
		assert.Equal(t, yerror.PoolOverflow, err, err)
	}
}

func TestCheckTxnSize(t *testing.T) {
	pool := initTxpool()
	pool.TxnMaxSize = 1
	err := pool.Insert(tx1)
	if err != nil {
		assert.Equal(t, yerror.TxnTooLarge, err, err)
	}
}

func TestPackFor(t *testing.T) {
	pool := initTxpool()
	err := pool.Insert(tx1)
	if err != nil {
		t.Fatalf("insert tx1 failed: %v", err)
	}
	err = pool.Insert(tx2)
	if err != nil {
		t.Fatalf("insert tx2 failed: %v", err)
	}
	err = pool.Insert(tx3)
	if err != nil {
		t.Fatalf("insert tx3 failed: %v", err)
	}

	txns, err := pool.PackFor(3, func(tx *types.SignedTxn) bool {
		return tx.Raw.Caller == caller2
	})
	if err != nil {
		t.Fatalf("pack txns failed: %v", err)
	}
	assert.Equal(t, txns[0], tx3)
}
