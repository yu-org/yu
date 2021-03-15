package txpool

import (
	. "yu/common"
	. "yu/txn"
)

type ItxPool interface {
	// return pool size of txpool
	PoolSize() uint64
	// txpool with the base check-functions
	WithBaseChecks(checkFns []TxnCheck) ItxPool
	// base check txn
	BaseCheck(IsignedTxn) error
	// tripods check the txn
	TripodsCheck(IsignedTxn) error
	// insert into txpool
	Insert(workerName string, txn IsignedTxn) error
	// batch insert into txpool
	BatchInsert(workerName string, txns SignedTxns) error
	// package some txns to send to tripods
	Package(workerName string, numLimit uint64) ([]IsignedTxn, error)
	// pacakge txns according to specific conditions
	PackageFor(workerName string, numLimit uint64, filter func(IsignedTxn) error) ([]IsignedTxn, error)
	// get txn content of txn-hash from p2p network
	SyncTxns(workerName string, hashes []Hash) error
	// remove txns after execute all tripods
	Remove(workerName string) error
}
