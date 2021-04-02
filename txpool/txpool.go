package txpool

import (
	. "yu/txn"
)

type ItxPool interface {
	NewEmptySignedTxn() *SignedTxn
	NewEmptySignedTxns() SignedTxns
	// return pool size of txpool
	PoolSize() uint64
	// txpool with the base check-functions
	WithBaseChecks(checkFns []TxnCheck) ItxPool
	// base check txn
	BaseCheck(*SignedTxn) error
	// tripods check the txn
	TripodsCheck(*SignedTxn) error
	// use for SyncTxns
	NecessaryCheck(stxn *SignedTxn) error
	// insert into txpool
	Insert(workerName string, txn *SignedTxn) error
	// batch insert into txpool
	BatchInsert(workerName string, txns SignedTxns) error
	// package some txns to send to tripods
	Package(workerName string, numLimit uint64) ([]*SignedTxn, error)
	// pacakge txns according to specific conditions
	PackageFor(workerName string, numLimit uint64, filter func(*SignedTxn) error) ([]*SignedTxn, error)
	// get txn content of txn-hash from p2p network
	// SyncTxns(hashes []Hash) error
	// remove txns after execute all tripods
	Flush() error
}
