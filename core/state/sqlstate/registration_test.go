package sqlstate

import (
	"path/filepath"
	"testing"
)

type inferredOrder struct {
	ID    string `gorm:"primaryKey"`
	Owner string `gorm:"index"`
}

type pointerNamedOrder inferredOrder

// TableName overrides only the business suffix, including for value registration.
func (*pointerNamedOrder) TableName() string { return "explicit_orders" }

type invalidNamedOrder inferredOrder

// TableName deliberately supplies an invalid identifier to test validation.
func (invalidNamedOrder) TableName() string { return "unsafe-name" }

// TestRegisterInfersNameAndCreatesTable checks initial creation, indexes,
// repeated registration and existing row preservation after reopening the store.
func TestRegisterInfersNameAndCreatesTable(t *testing.T) {
	s, db := openTestStore(t, filepath.Join(t.TempDir(), "registration.db"))
	const name = "orderbook_inferred_orders"
	if db.Migrator().HasTable(name) {
		t.Fatal("test table already exists")
	}
	table, err := s.RegisterTable("orderbook", &inferredOrder{})
	must(t, err)
	if table.Name() != name || !db.Migrator().HasTable(name) {
		t.Fatalf("inferred table was not created: %s", table.Name())
	}
	if !db.Migrator().HasIndex(name, "idx_orderbook_inferred_orders_owner") {
		t.Fatal("declared secondary index was not created")
	}
	again, err := s.RegisterTable("orderbook", inferredOrder{})
	must(t, err)
	if again != table {
		t.Fatal("matching registration did not reuse the handle")
	}
	must(t, s.Begin("genesis", "", 0))
	must(t, table.Put(inferredOrder{ID: "one", Owner: "alice"}))
	commitTest(t, s)
	must(t, s.Finalize("genesis"))
	// A new Store stands in for restart; only the new owner is used from here.
	reopened, err := Open(db)
	must(t, err)
	table, err = reopened.RegisterTable("orderbook", &inferredOrder{})
	must(t, err)
	var row inferredOrder
	must(t, table.GetFinalized("one", &row))
	if row.Owner != "alice" {
		t.Fatal("registration changed the existing state row")
	}
	var rows []inferredOrder
	must(t, table.FindFinalized(map[string]any{"owner": "alice"}, &rows))
	if len(rows) != 1 || rows[0].ID != "one" {
		t.Fatal("registration changed the existing SQL rows")
	}
}

// TestRegisterHonorsPointerTableName checks model-derived names stay namespaced.
func TestRegisterHonorsPointerTableName(t *testing.T) {
	s, db := openTestStore(t, filepath.Join(t.TempDir(), "named.db"))
	table, err := s.RegisterTable("orderbook", pointerNamedOrder{})
	must(t, err)
	if table.Name() != "orderbook_explicit_orders" || !db.Migrator().HasTable(table.Name()) {
		t.Fatalf("model TableName was not honored: %s", table.Name())
	}
	if db.Migrator().HasTable("explicit_orders") {
		t.Fatal("created an unnamespaced model table")
	}
	if _, err := s.RegisterTable("orderbook", &invalidNamedOrder{}); err == nil {
		t.Fatal("invalid model table name accepted")
	}
}

// TestRegisterRejectsPersistedSchemaConflict checks failure precedes AutoMigrate.
func TestRegisterRejectsPersistedSchemaConflict(t *testing.T) {
	s, table, db := testStore(t)
	seedGenesis(t, s, table)
	reopened, err := Open(db)
	must(t, err)
	if _, err := reopened.RegisterTable("orderbook", &conflictingOrder{}); err == nil {
		t.Fatal("persisted schema conflict accepted")
	}
	if db.Migrator().HasColumn(table.Name(), "owner") {
		t.Fatal("conflicting registration altered existing SQL schema")
	}
	table, err = reopened.RegisterTable("orderbook", &testOrder{})
	must(t, err)
	var row testOrder
	must(t, table.GetFinalized("a", &row))
	if row.Status != "open" || row.Amount != 10 {
		t.Fatal("conflicting registration changed finalized state")
	}
}
