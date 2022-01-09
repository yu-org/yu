package txpool

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

type ItxPool interface {
	// PoolSize return pool size of txpool
	PoolSize() uint64

	WithBaseChecks(checkFns []TxnCheck) ItxPool
	WithTripodChecks(checkFns []TxnCheck) ItxPool

	BaseCheck(*SignedTxn) error
	TripodsCheck(stxn *SignedTxn) error

	// NecessaryCheck uses for SyncTxns
	NecessaryCheck(stxn *SignedTxn) error

	Insert(txn *SignedTxn) error
	BatchInsert(txns SignedTxns) []error

	// Pack packs some txns to send to tripods
	Pack(numLimit uint64) ([]*SignedTxn, error)

	PackFor(numLimit uint64, filter func(txn *SignedTxn) bool) ([]*SignedTxn, error)

	// GetTxn returns unpacked txn
	GetTxn(hash Hash) (*SignedTxn, error)
	// Reset deletes packed txns
	Reset(*CompactBlock) error
}
