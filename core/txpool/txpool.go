package txpool

import (
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types"
)

type ItxPool interface {
	//NewEmptySignedTxn() *SignedTxn
	//NewEmptySignedTxns() SignedTxns

	// return pool size of txpool
	PoolSize() uint64
	// txpool with the check-functions
	WithBaseChecks(checkFns []TxnCheck) ItxPool
	WithTripodChecks(checkFns []TxnCheck) ItxPool
	// base check txn
	BaseCheck(*types.SignedTxn) error
	TripodsCheck(stxn *types.SignedTxn) error
	// use for SyncTxns
	NecessaryCheck(stxn *types.SignedTxn) error
	// insert into txpool
	Insert(txn *types.SignedTxn) error
	// batch insert into txpool
	BatchInsert(txns types.SignedTxns) error
	// package some txns to send to tripods
	Pack(numLimit uint64) ([]*types.SignedTxn, error)
	// pacakge txns according to specific conditions
	PackFor(numLimit uint64, filter func(*types.SignedTxn) error) ([]*types.SignedTxn, error)

	GetTxn(hash Hash) (*types.SignedTxn, error)

	RemoveTxns(hashes []Hash) error
	// remove txns after execute all tripods
	Reset() error
}
