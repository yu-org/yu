package txpool

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

type ItxPool interface {
	// Capacity return pool capacity of txpool
	Capacity() int
	// Size return pool size of txpool
	Size() int

	WithBaseCheck(checkFn TxnChecker) ItxPool
	WithTripodCheck(tripodName string, checker TxnChecker) ItxPool

	BaseCheck(*SignedTxn) error
	TripodsCheck(stxn *SignedTxn) error

	// NecessaryCheck uses for SyncTxns
	NecessaryCheck(stxn *SignedTxn) error

	Exist(txnHash Hash) bool
	CheckTxn(stxn *SignedTxn) error

	Insert(txn *SignedTxn) error
	SetOrder(order map[int]Hash)
	SortTxns(fn func(txns []*SignedTxn) []*SignedTxn)

	// Pack packs some txns to send to tripods
	Pack(numLimit uint64) ([]*SignedTxn, error)

	PackFor(numLimit uint64, filter func(txn *SignedTxn) bool) ([]*SignedTxn, error)

	// GetTxn returns unpacked txn
	GetTxn(hash Hash) (*SignedTxn, error)
	GetAllTxns() ([]*SignedTxn, error)
	// Reset Deletes packed txns
	Reset(txns SignedTxns) error
	ResetByHashes(hashes []Hash) error
}

type IunpackedTxns interface {
	Insert(input *SignedTxn)
	Deletes(txnHashes []Hash)
	Exist(txnHash Hash) bool
	Get(txnHash Hash) *SignedTxn
	GetAll() []*SignedTxn
	Gets(numLimit uint64, filter func(txn *SignedTxn) bool) []*SignedTxn
	SortTxns(fn func(txns []*SignedTxn) []*SignedTxn)
	SetOrder(order map[int]Hash)
	Size() int
}
