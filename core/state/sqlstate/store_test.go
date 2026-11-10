package sqlstate

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testOrder struct {
	ID     string `gorm:"primaryKey"`
	Status string `gorm:"index"`
	Amount uint64
}

type testCash struct {
	ID    string `gorm:"primaryKey"`
	Owner string `gorm:"index"`
}

// TableName pins the fixture name independently of the Go type name.
func (testOrder) TableName() string { return "orders" }

// TableName pins the fixture cash table name.
func (testCash) TableName() string { return "cash" }

type conflictingOrder testCash

// TableName deliberately collides with the order fixture to test schema rejection.
func (conflictingOrder) TableName() string { return "orders" }

type suffixC testCash

// TableName selects the first suffix in an ambiguous namespace pair.
func (suffixC) TableName() string { return "c" }

type suffixBC testCash

// TableName selects the second suffix in an ambiguous namespace pair.
func (suffixBC) TableName() string { return "b_c" }

// openTestStore opens an isolated SQLite database through the generic GORM API.
func openTestStore(t *testing.T, path string) (*Store, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = raw.Close() })
	store, err := Open(db)
	if err != nil {
		t.Fatal(err)
	}
	return store, db
}

// testStore creates a fresh state store and its registered order table.
func testStore(t *testing.T) (*Store, *Table, *gorm.DB) {
	t.Helper()
	s, db := openTestStore(t, filepath.Join(t.TempDir(), "state.db"))
	table, err := s.RegisterTable("orderbook", &testOrder{})
	if err != nil {
		t.Fatal(err)
	}
	return s, table, db
}

// must performs a required setup action.
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// commitTest commits an execution and returns its authenticated root.
func commitTest(t *testing.T, s *Store) []byte {
	t.Helper()
	root, err := s.Commit()
	must(t, err)
	return root
}

// seedGenesis writes and finalizes a known initial order.
func seedGenesis(t *testing.T, s *Store, table *Table) {
	t.Helper()
	must(t, s.Begin("genesis", "", 0))
	must(t, table.Put(testOrder{ID: "a", Status: "open", Amount: 10}))
	commitTest(t, s)
	must(t, s.Finalize("genesis"))
}

// TestTransactionRollback checks that failed writes cannot leak into later successes.
func TestTransactionRollback(t *testing.T) {
	s, table, _ := testStore(t)
	seedGenesis(t, s, table)
	must(t, s.Begin("b1", "genesis", 1))
	must(t, table.Put(testOrder{ID: "a", Status: "matched", Amount: 10}))
	must(t, s.NextTxn())
	must(t, table.Delete("a"))
	must(t, table.Put(testOrder{ID: "failed", Status: "open"}))
	must(t, s.DiscardTxn())
	must(t, table.Put(testOrder{ID: "good", Status: "open"}))
	commitTest(t, s)
	var row testOrder
	must(t, table.Get("a", &row))
	if row.Status != "matched" {
		t.Fatalf("earlier transaction lost: %+v", row)
	}
	if !errors.Is(table.Get("failed", &row), ErrNotFound) {
		t.Fatal("failed transaction survived")
	}
	must(t, table.Get("good", &row))
	must(t, table.GetFinalized("a", &row))
	if row.Status != "open" {
		t.Fatal("pending writes reached finalized state")
	}
}

// TestForkIsolationAndFinalization checks competing values, deletion and retained descendants.
func TestForkIsolationAndFinalization(t *testing.T) {
	s, table, db := testStore(t)
	seedGenesis(t, s, table)
	must(t, s.Begin("a1", "genesis", 1))
	must(t, table.Put(testOrder{ID: "a", Status: "matched"}))
	commitTest(t, s)
	must(t, s.Begin("a2", "a1", 2))
	must(t, table.Delete("a"))
	commitTest(t, s)
	must(t, s.Begin("b1", "genesis", 1))
	var row testOrder
	must(t, table.Get("a", &row))
	if row.Status != "open" {
		t.Fatal("rival branch leaked into parent read")
	}
	must(t, table.Put(testOrder{ID: "a", Status: "rival"}))
	commitTest(t, s)
	must(t, table.GetPending("a1", "a", &row))
	if row.Status != "matched" {
		t.Fatal("competing version overwrote row")
	}
	if !errors.Is(table.GetPending("a2", "a", &row), ErrNotFound) {
		t.Fatal("tombstone missing")
	}
	must(t, s.Finalize("a1"))
	must(t, table.GetFinalized("a", &row))
	if row.Status != "matched" {
		t.Fatal("wrong finalized value")
	}
	for _, old := range []string{"genesis", "b1"} {
		if _, err := s.Root(old); !errors.Is(err, ErrVersionUnavailable) {
			t.Fatalf("retired version %s accessible: %v", old, err)
		}
	}
	if !errors.Is(table.GetPending("a2", "a", &row), ErrNotFound) {
		t.Fatal("pending descendant broken by GC")
	}
	must(t, s.Finalize("a2"))
	if !errors.Is(table.GetFinalized("a", &row), ErrNotFound) {
		t.Fatal("finalized delete missing")
	}
	var n int64
	must(t, db.Table(table.Name()).Count(&n).Error)
	if n != 0 {
		t.Fatal("materialized SQL row was not deleted")
	}
	must(t, db.Model(&storedRow{}).Count(&n).Error)
	if n != 0 {
		t.Fatal("unreachable row values were not collected")
	}
}

// TestRestartKeepsPendingBranches checks persisted roots and re-registration after reopening.
func TestRestartKeepsPendingBranches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	s, db := openTestStore(t, path)
	table, err := s.RegisterTable("orderbook", &testOrder{})
	must(t, err)
	seedGenesis(t, s, table)
	must(t, s.Begin("pending", "genesis", 1))
	must(t, table.Put(testOrder{ID: "a", Status: "pending"}))
	root := commitTest(t, s)
	raw, err := db.DB()
	must(t, err)
	must(t, raw.Close())
	s, _ = openTestStore(t, path)
	table, err = s.RegisterTable("orderbook", &testOrder{})
	must(t, err)
	got, err := s.Root("pending")
	must(t, err)
	if !bytes.Equal(root, got) {
		t.Fatal("root changed on restart")
	}
	var row testOrder
	must(t, table.GetPending("pending", "a", &row))
	if row.Status != "pending" {
		t.Fatal("pending value lost on restart")
	}
	must(t, s.Finalize("pending"))
	must(t, table.GetFinalized("a", &row))
}

// TestRootIsFullStateAndIndependentOfWriteOrder compares incremental and rebuilt states.
func TestRootIsFullStateAndIndependentOfWriteOrder(t *testing.T) {
	s, table, _ := testStore(t)
	seedGenesis(t, s, table)
	must(t, s.Begin("next", "genesis", 1))
	must(t, table.Put(testOrder{ID: "b", Status: "new"}))
	root := commitTest(t, s)
	other, otherTable, _ := testStore(t)
	must(t, other.Begin("rebuilt", "", 0))
	must(t, otherTable.Put(testOrder{ID: "b", Status: "new"}))
	must(t, otherTable.Put(testOrder{ID: "a", Status: "open", Amount: 10}))
	if !bytes.Equal(root, commitTest(t, other)) {
		t.Fatal("root depends on insertion order or excludes unchanged rows")
	}
}

// TestRootMismatchRollsBackSQL checks the entire failed commit has no persisted side effects.
func TestRootMismatchRollsBackSQL(t *testing.T) {
	s, table, db := testStore(t)
	must(t, s.Begin("bad", "", 0))
	must(t, table.Put(testOrder{ID: "a"}))
	if _, err := s.CommitExpected(make([]byte, 32)); err == nil {
		t.Fatal("accepted wrong root")
	}
	for _, model := range []any{&storedNode{}, &storedRow{}, &pendingChange{}, &versionRecord{}} {
		var n int64
		must(t, db.Model(model).Count(&n).Error)
		if n != 0 {
			t.Fatalf("failed commit persisted %T", model)
		}
	}
	commitTest(t, s)
}

// TestCrossTableFinalizeIsAtomic checks a failed projection cannot partially finalize.
func TestCrossTableFinalizeIsAtomic(t *testing.T) {
	s, orders, db := testStore(t)
	cash, err := s.RegisterTable("account", &testCash{})
	must(t, err)
	must(t, s.Begin("genesis", "", 0))
	must(t, orders.Put(testOrder{ID: "a"}))
	must(t, cash.Put(testCash{ID: "cash", Owner: "alice"}))
	commitTest(t, s)
	delete(s.tables, cash.Name())
	if err := s.Finalize("genesis"); err == nil {
		t.Fatal("finalized with missing table registration")
	}
	for _, name := range []string{orders.Name(), cash.Name()} {
		var n int64
		must(t, db.Table(name).Count(&n).Error)
		if n != 0 {
			t.Fatal("partial table promotion")
		}
	}
	head, err := s.Finalized()
	must(t, err)
	if head != "" {
		t.Fatal("failed finalization advanced head")
	}
	s.tables[cash.Name()] = cash
	must(t, s.Finalize("genesis"))
	var rows []testCash
	must(t, cash.FindFinalized(map[string]any{"owner": "alice"}, &rows))
	if len(rows) != 1 {
		t.Fatal("indexed finalized query missing row")
	}
}

// TestAbortAndDropKeepFinalizedState checks execution and fork cleanup boundaries.
func TestAbortAndDropKeepFinalizedState(t *testing.T) {
	s, table, _ := testStore(t)
	seedGenesis(t, s, table)
	must(t, s.Begin("aborted", "genesis", 1))
	must(t, table.Delete("a"))
	if !errors.Is(s.Finalize("genesis"), ErrActiveExecution) {
		t.Fatal("finalized during active execution")
	}
	s.Abort()
	must(t, s.Begin("pending", "genesis", 1))
	must(t, table.Put(testOrder{ID: "a", Status: "new"}))
	commitTest(t, s)
	must(t, s.DropPending("pending"))
	var row testOrder
	must(t, table.GetFinalized("a", &row))
	if row.Status != "open" {
		t.Fatal("drop corrupted finalized row")
	}
	if s.DropPending("genesis") == nil {
		t.Fatal("dropped finalized state")
	}
}

// TestTableRegistrationRejectsNameAndSchemaConflicts enforces stable table identities.
func TestTableRegistrationRejectsNameAndSchemaConflicts(t *testing.T) {
	s, _, _ := testStore(t)
	if _, err := s.RegisterTable("bad-name", &testOrder{}); err == nil {
		t.Fatal("unsafe identifier accepted")
	}
	if _, err := s.RegisterTable("orderbook", &conflictingOrder{}); err == nil {
		t.Fatal("schema conflict accepted")
	}
	_, err := s.RegisterTable("a_b", &suffixC{})
	must(t, err)
	if _, err := s.RegisterTable("a", &suffixBC{}); err == nil {
		t.Fatal("ambiguous namespace accepted")
	}
}
