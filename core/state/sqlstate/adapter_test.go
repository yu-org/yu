package sqlstate

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types"
)

type testTripod string

// Name returns the compatibility namespace.
func (t testTripod) Name() string { return string(t) }

// TestAdapterLifecycle exercises the actual IState adapter and sealed hash binding.
func TestAdapterLifecycle(t *testing.T) {
	_, db := openTestStore(t, filepath.Join(t.TempDir(), "adapter.db"))
	s, err := NewState(db)
	must(t, err)
	must(t, s.PrepareTripod("ledger"))
	genesis := &types.Block{Header: &types.Header{Hash: common.HexToHash("01")}}
	must(t, s.InitializeState(genesis))
	block := &types.Block{Header: &types.Header{PrevHash: genesis.Hash, Height: 1}}
	s.StartBlock(block)
	s.Set(testTripod("ledger"), []byte("a"), []byte("good"))
	s.NextTxn()
	s.Set(testTripod("ledger"), []byte("a"), []byte("failed"))
	s.Discard()
	root, err := s.Commit()
	must(t, err)
	block.StateRoot = common.BytesToHash(root)
	block.Hash, err = types.ComputeHeaderHash(block)
	must(t, err)
	must(t, s.BindBlock(block))
	if !s.ExecutionCommitted(block) {
		t.Fatal("block was not bound")
	}
	got, err := s.Get(testTripod("ledger"), []byte("a"))
	must(t, err)
	if !bytes.Equal(got, []byte("good")) {
		t.Fatal("failed transaction survived")
	}
	must(t, s.FinalizeState(block))
	got, err = s.GetFinalized(testTripod("ledger"), []byte("a"))
	must(t, err)
	if !bytes.Equal(got, []byte("good")) {
		t.Fatal("finalized KV value lost")
	}
	_, err = s.GetByBlockHash(testTripod("ledger"), []byte("a"), genesis)
	if !errors.Is(err, ErrVersionUnavailable) {
		t.Fatalf("old finalized state retained: %v", err)
	}
	must(t, s.InitializeState(genesis))
	wrongGenesis := &types.Block{Header: &types.Header{Hash: common.HexToHash("02")}}
	if s.InitializeState(wrongGenesis) == nil {
		t.Fatal("accepted another chain's database")
	}
}

// TestAdapterAbortRemovesBoundPendingState checks an append failure is recoverable.
func TestAdapterAbortRemovesBoundPendingState(t *testing.T) {
	_, db := openTestStore(t, filepath.Join(t.TempDir(), "abort.db"))
	s, err := NewState(db)
	must(t, err)
	must(t, s.PrepareTripod("ledger"))
	genesis := &types.Block{Header: &types.Header{Hash: common.HexToHash("01")}}
	must(t, s.InitializeState(genesis))
	block := &types.Block{Header: &types.Header{PrevHash: genesis.Hash, Height: 1}}
	must(t, s.BeginExecution(block))
	root, err := s.Commit()
	must(t, err)
	block.StateRoot = common.BytesToHash(root)
	block.Hash, err = types.ComputeHeaderHash(block)
	must(t, err)
	must(t, s.BindBlock(block))
	s.AbortExecution()
	must(t, s.ExecutionError())
	if _, err := s.Root(block.Hash.String()); !errors.Is(err, ErrVersionUnavailable) {
		t.Fatal("aborted version remained")
	}
	must(t, s.BeginExecution(block))
	s.AbortExecution()
}

// TestScanMergesOnlyTheSelectedView checks created, updated and deleted pending rows.
func TestScanMergesOnlyTheSelectedView(t *testing.T) {
	s, table, _ := testStore(t)
	seedGenesis(t, s, table)
	must(t, s.Begin("pending", "genesis", 1))
	must(t, table.Put(testOrder{ID: "b", Status: "new"}))
	must(t, table.Delete("a"))
	var rows []testOrder
	must(t, table.Scan(&rows))
	if len(rows) != 1 || rows[0].ID != "b" {
		t.Fatalf("wrong pending scan: %+v", rows)
	}
	must(t, s.DiscardTxn())
	must(t, table.Scan(&rows))
	if len(rows) != 1 || rows[0].ID != "a" {
		t.Fatal("scan ignored rollback")
	}
	s.Abort()
}

// TestCorruptRowsFailClosed rejects a valid node pointing to modified row bytes.
func TestCorruptRowsFailClosed(t *testing.T) {
	s, table, db := testStore(t)
	seedGenesis(t, s, table)
	var row storedRow
	must(t, db.First(&row).Error)
	must(t, db.Model(&storedRow{}).Where("hash = ?", row.Hash).Update("data", []byte("tampered")).Error)
	var order testOrder
	if err := table.GetFinalized("a", &order); err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("corruption hidden as absence: %v", err)
	}
}
