package kernel

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/env"
	"github.com/yu-org/yu/core/state/sqlstate"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type receiptMemory struct {
	memTxDB
	receipts map[common.Hash]*types.Receipt
}

// SetReceipts retains final metadata for execution integration tests.
func (m *receiptMemory) SetReceipts(receipts map[common.Hash]*types.Receipt) error {
	m.receipts = receipts
	return nil
}

type sqlWriter struct{ *tripod.Tripod }

// Write deliberately fails after staging a write when requested.
func (w *sqlWriter) Write(ctx *context.WriteContext) error {
	key, err := ctx.TryGetString("key")
	if err != nil {
		return err
	}
	w.Set([]byte(key), []byte("value"))
	if key == "failed" {
		return errors.New("deliberate transaction failure")
	}
	return nil
}

// newSQLTestState opens an isolated SQL backend without anchoring genesis.
func newSQLTestState(t *testing.T) *sqlstate.State {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "state.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	s, err := sqlstate.NewState(db)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// sqlTestKernel constructs an execution kernel without starting networking.
func sqlTestKernel(t *testing.T) (*Kernel, *sqlstate.State, *sqlWriter, *receiptMemory, common.Hash) {
	t.Helper()
	s := newSQLTestState(t)
	if err := s.PrepareTripod("ledger"); err != nil {
		t.Fatal(err)
	}
	genesis := &types.Block{Header: &types.Header{Hash: common.HexToHash("01")}}
	if err := s.InitializeState(genesis); err != nil {
		t.Fatal(err)
	}
	memory := &receiptMemory{}
	chainEnv := &env.ChainEnv{State: s, TxDB: memory}
	w := &sqlWriter{Tripod: tripod.NewTripodWithName("ledger")}
	w.SetChainEnv(chainEnv)
	w.SetWritings(w.Write)
	land := tripod.NewLand()
	land.RegisterTripods(w.Tripod)
	k := &Kernel{ChainEnv: chainEnv, Land: land}
	return k, s, w, memory, genesis.Hash
}

// executionTxn builds an already-validated transaction for executor tests.
func executionTxn(key string) *types.SignedTxn {
	return &types.SignedTxn{
		Raw:     &types.UnsignedTxn{WrCall: &common.WrCall{TripodName: "ledger", FuncName: "Write", Params: `{"key":"` + key + `"}`}},
		TxnHash: common.BytesToHash(common.Sha256([]byte(key))), Address: make([]byte, 20),
	}
}

// TestSQLExecutionReplayAndRollback covers roots, failed writes and sealed receipt metadata.
func TestSQLExecutionReplayAndRollback(t *testing.T) {
	k, s, writer, receipts, genesis := sqlTestKernel(t)
	block := &types.Block{Header: &types.Header{PrevHash: genesis, Height: 1, LeiLimit: 100}, Txns: types.SignedTxns{executionTxn("good"), executionTxn("failed")}}
	var err error
	block.TxnRoot, err = types.MakeTxnRoot(block.Txns)
	if err != nil {
		t.Fatal(err)
	}
	if err := k.SeqExecuteWritings(block); err != nil {
		t.Fatal(err)
	}
	if receipts.receipts != nil {
		t.Fatal("receipts persisted before sealing")
	}
	value, err := writer.Get([]byte("good"))
	if err != nil || string(value) != "value" {
		t.Fatalf("good write missing: %s %v", value, err)
	}
	value, err = writer.Get([]byte("failed"))
	if err != nil || value != nil {
		t.Fatal("failed transaction changed state")
	}
	block.Hash, err = types.ComputeHeaderHash(block)
	if err != nil {
		t.Fatal(err)
	}
	if err := k.BindState(block); err != nil {
		t.Fatal(err)
	}
	for _, receipt := range receipts.receipts {
		if receipt.BlockHash != block.Hash {
			t.Fatal("receipt has provisional hash")
		}
	}
	if receipts.receipts[executionTxn("failed").TxnHash].Error == "" {
		t.Fatal("failed receipt marked successful")
	}
	encoded, err := block.Encode()
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := types.DecodeBlock(encoded)
	if err != nil {
		t.Fatal(err)
	}
	other, receiver, _, _, _ := sqlTestKernel(t)
	if err := other.SeqExecuteWritings(replayed); err != nil {
		t.Fatalf("valid replay failed: %v", err)
	}
	if err := other.BindState(replayed); err != nil {
		t.Fatal(err)
	}
	if err := receiver.FinalizeState(replayed); err != nil {
		t.Fatal(err)
	}
	if err := s.FinalizeState(block); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Root(genesis.String()); !errors.Is(err, sqlstate.ErrVersionUnavailable) {
		t.Fatal("old finalized root survived")
	}
}

// TestSQLRejectsWrongStateRoot ensures failed block verification leaves no active execution.
func TestSQLRejectsWrongStateRoot(t *testing.T) {
	k, s, _, _, genesis := sqlTestKernel(t)
	block := &types.Block{Header: &types.Header{PrevHash: genesis, Height: 1, LeiLimit: 100, StateRoot: common.HexToHash("ff")}}
	if k.SeqExecuteWritings(block) == nil {
		t.Fatal("accepted wrong state root")
	}
	if err := s.ExecutionError(); err != nil {
		t.Fatalf("failed execution did not abort cleanly: %v", err)
	}
	block.StateRoot = common.NullHash
	if err := k.SeqExecuteWritings(block); err != nil {
		t.Fatalf("valid retry failed: %v", err)
	}
	s.AbortExecution()
}

// TestSQLFinalityRecovery repairs the crash window between state and chain markers.
func TestSQLFinalityRecovery(t *testing.T) {
	k := newGenesisKernel(t)
	state := newSQLTestState(t)
	k.State = state
	if err := k.initGenesis(); err != nil {
		t.Fatal(err)
	}
	genesis, err := k.Chain.GetGenesis()
	if err != nil {
		t.Fatal(err)
	}
	if genesis.StateRoot == common.NullHash {
		t.Fatal("default SQL genesis does not commit to empty state")
	}
	if err := state.InitializeState(genesis); err != nil {
		t.Fatal(err)
	}
	k.State = state
	block := &types.Block{Header: &types.Header{PrevHash: genesis.Hash, Height: 1, LeiLimit: 100}}
	if err := k.SeqExecuteWritings(block); err != nil {
		t.Fatal(err)
	}
	block.Hash, err = types.ComputeHeaderHash(block)
	if err != nil {
		t.Fatal(err)
	}
	if err := k.BindState(block); err != nil {
		t.Fatal(err)
	}
	if err := k.Chain.AppendBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := state.FinalizeState(block); err != nil {
		t.Fatal(err)
	}
	if err := k.reconcileStateFinality(); err != nil {
		t.Fatal(err)
	}
	last, err := k.Chain.LastFinalized()
	if err != nil {
		t.Fatal(err)
	}
	if last.Hash != block.Hash {
		t.Fatal("chain marker was not repaired")
	}
}
