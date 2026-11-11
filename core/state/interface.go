package state

import (
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
)

// TODO: need to add prove() and verify()
type IState interface {
	Set(triName NameString, key, value []byte)
	Delete(triName NameString, key []byte)
	Get(triName NameString, key []byte) ([]byte, error)
	GetFinalized(triName NameString, key []byte) ([]byte, error)
	Exist(triName NameString, key []byte) bool
	GetByBlockHash(triName NameString, key []byte, block *types.Block) ([]byte, error)
	Commit() ([]byte, error)
	NextTxn()
	Discard()
	DiscardAll()
	StartBlock(block *types.Block)
	FinalizeBlock(block *types.Block)
}

// ExecutionState is implemented by backends with durable pending block versions.
// Consensus must bind state after sealing and before appending a block.
type ExecutionState interface {
	BeginExecution(block *types.Block) error
	ExecutionError() error
	ExecutionCommitted(block *types.Block) bool
	AbortExecution()
	BindBlock(block *types.Block) error
}

// InitializableState anchors a backend to the chain's actual genesis identity.
type InitializableState interface {
	InitializeState(genesis *types.Block) error
	GenesisStateRoot() common.Hash
}

// PreparedState initializes compatibility tables before transactions execute.
type PreparedState interface{ PrepareTripod(name string) error }

// FinalizingState exposes errors without the legacy void finalization callback.
type FinalizingState interface {
	FinalizeState(block *types.Block) error
}

// FinalizedStateHead exposes the authoritative durable state boundary for recovery.
type FinalizedStateHead interface {
	FinalizedStateHead() (common.Hash, common.Hash, error)
}

// NewStateDB constructs a legacy KV backend; SQL backends need a SQL connection.
func NewStateDB(typ string, kvdb kv.Kvdb) IState {
	// FIXME
	switch typ {
	case "sql":
		panic("SQL state requires a SQL connection; use startup or sqlstate.NewState")
	case "no":
		return new(NoStateDB)
	default:
		return NewSpmtKV(nil, kvdb)
	}
}
