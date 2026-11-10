# SQL table state

This backend uses a single SQL database for finalized business tables, pending
changes, immutable Merkle nodes, row content and retained block versions. It
accepts an existing GORM connection. SQLite is the default; MySQL/InnoDB and
PostgreSQL use the existing SQL driver adapters. A database has one execution
owner, not multiple independently running nodes.

## Configuration

```toml
statedb_type = "sql"

[state_sql]
sql_db_type = "sqlite"
dsn = "state.db"
```

SQLite relative paths are resolved below `data_dir`. MySQL accepts `mysql`;
PostgreSQL accepts `postgre`, `postgres` or `postgresql` and a driver DSN.
`config.InitDefaultCfg()` selects SQL state for new chains. Existing configurations
without this selection retain the legacy backend. This is a new state and block
commitment format: existing KV state databases are not automatically migrated.

## Tables

Register models before executing or synchronizing blocks, for example in a
business tripod's `InitChain` before the synchronizer initializes:

```go
type Order struct {
    ID     string `gorm:"primaryKey"`
    Owner  string `gorm:"index"`
    Status string `gorm:"index"`
    Amount int64
}

// InitChain registers this tripod's business table before executing blocks.
func (o *OrderBook) InitChain() {
    table, err := o.RegisterStateTable("orders", &Order{})
    if err != nil {
        panic(err)
    }
    o.orders = table
}
```

The physical table is `orderbook_orders`, regardless of a model's `TableName`
method. Names must be lowercase identifiers, with a combined maximum of 63 bytes.
Conflicting registrations, including ambiguous concatenated names, are rejected.
Composite primary keys are supported; pass a model holding its primary key
fields to lookups. Numeric primary keys need `autoIncrement:false`.

`Put`, `Get` and `Delete` operate on the active transaction/pending view. `Scan`
returns a complete pending view in deterministic encoded-key order; callers can
filter it for matching. `FindFinalized` uses real SQL columns and their indexes,
but deliberately queries only the latest finalized table. `GetFinalized` reads
the latest authenticated finalized view. Rows must be supplied completely.

Only persisted fields participate in the canonical encoding, sorted by SQL
column name. Fields support strings, booleans, integers, bytes and nullable
pointers to these types. Encode richer values into explicit string/byte columns.
Floating-point and native time columns are excluded to avoid SQL precision differences;
use integer timestamps. Indexed strings are limited to 191 bytes for portable keys.
Associations, SQL type overrides, generated IDs/timestamps/defaults, custom GORM
serializers, check constraints and secondary uniqueness constraints are rejected.
Business validation, including uniqueness checks against pending data, belongs
in the transaction. Integer values must fit the portable signed SQL range;
larger integers can use decimal strings. Primary keys must be strings or integers;
string keys cannot end in spaces. Model types, tags and encodings are protocol
configuration. Registration refuses schema changes without an explicit migration.
Direct SQL modifications, including hooks and updates outside `Table`, bypass
the commitment and are unsupported.

## Commitments and retention

All tables share one compressed, 256-bit-path sparse Merkle tree:

- Paths hash an unambiguous `(physical table, encoded primary key)` tuple.
- Row blobs encode the table, primary key and canonical row content.
- Leaves commit to a path and a content-addressed row hash.
- Branches commit to ordered child hashes; distinct domain tags separate rows,
  leaves, branches, paths and the empty subtree.
- Lone terminal leaves compress empty paths. Changes copy affected paths;
  no update overwrites a retained node or row.

The store retains the latest finalized version and pending descendants. Advancing
finality atomically materializes the chosen changes, updates the finalized pointer,
retires previous finalized versions and competing forks, and sweeps unreachable
nodes and rows. Finalized history is intentionally unavailable. Shared content
still needed by pending descendants survives. Collection is currently synchronous
mark-and-sweep; pending scans do not yet use SQL secondary indexes.

## Execution integration

`sqlstate.State` implements `state.IState`. Legacy `Tripod.Set/Get` users get a
`tripod_kv` compatibility table; concrete tables use `RegisterStateTable`.

The kernel starts each execution from its actual parent, handles transaction
frames, checks a supplied StateRoot inside the SQL commit transaction, and aborts
failed executions. Commit first uses a local execution ID, allowing StateRoot to
be computed before the final block hash exists. Consensus must then seal the
header and call `ChainEnv.BindState(block)` before append. This binds the version
to the block hash and persists receipts with final metadata. Block-generated
business values must not depend on the final block hash if they enter StateRoot.

PoA follows this lifecycle and executes a produced block once before signing.
Receivers recompute roots. SQL-state receipt commitments exclude local save times
and the final block hash, and use deterministic ordering and domain-separated
hashes. Other consensus implementations must adopt the same lifecycle explicitly;
the legacy KV and no-state backends keep their prior behavior.

History synchronization executes, verifies and binds before appending. Receiving
a block alone does not prove finality; the consensus implementation must call
`FinalizeState` once it has made that decision. It returns errors, unlike the
legacy void `FinalizeBlock` callback. `DropPending` discards a losing branch and
its descendants; `AbortExecution` discards a failed current execution.

Genesis is anchored before tripod initialization. Custom genesis allocations
must be initialized and their models registered through `Store` before startup,
using the actual genesis identity and matching root. The genesis identity remains
recorded independently of pruned versions. On restart, the kernel repairs a chain
finality marker lagging the durable SQL state head, verifying block hash ancestry
and state root rather than selecting a block by height. SQL state writes are
atomic internally; block storage and receipt storage remain separate subsystems.

## Verification

Tests cover incremental roots against independent complete rebuilds, transaction
rollback, fork isolation, deletion, deterministic pending scans, restart, finalized
history retirement, descendant retention, garbage collection, cross-table
promotion rollback, corrupt rows, root mismatch rollback, execution replay,
receipt metadata, finality-marker recovery and three-node PoA replay.
SQLite is exercised with a real database. MySQL and PostgreSQL require live-service
conformance testing in the deployment environment; these tests do not claim it.
