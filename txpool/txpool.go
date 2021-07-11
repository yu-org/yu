package txpool

import (
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/txn"
)

type ItxPool interface {
	NewEmptySignedTxn() *SignedTxn
	NewEmptySignedTxns() SignedTxns
	// return pool size of txpool
	PoolSize() uint64
	// txpool with the check-functions
	WithBaseChecks(checkFns []TxnCheck) ItxPool
	WithTripodChecks(checkFns []TxnCheck) ItxPool
	// base check txn
	BaseCheck(*SignedTxn) error
	TripodsCheck(stxn *SignedTxn) error
	// use for SyncTxns
	NecessaryCheck(stxn *SignedTxn) error
	// insert into txpool
	Insert(workerName string, txn *SignedTxn) error
	// batch insert into txpool
	BatchInsert(workerName string, txns SignedTxns) error
	// package some txns to send to tripods
	Pack(workerName string, numLimit uint64) ([]*SignedTxn, error)
	// pacakge txns according to specific conditions
	PackFor(workerName string, numLimit uint64, filter func(*SignedTxn) error) ([]*SignedTxn, error)

	GetTxn(hash Hash) (*SignedTxn, error)

	RemoveTxns(hashes []Hash) error
	// remove txns after execute all tripods
	Flush() error

	Reset()
}
