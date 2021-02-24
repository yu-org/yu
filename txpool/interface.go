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
	AddTripodsCheck(checkFn TxnCheck)
	// insert into txCache for pending
	Pend(IsignedTxn) error
	// insert into txPool for tripods
	Insert(BlockNum, IsignedTxn) error
	// package some txns to send to tripods
	Package(numLimit uint64) ([]IsignedTxn, error)
	// pacakge txns according to specific conditions
	PackageFor(numLimit uint64, filter func(IsignedTxn) error) ([]IsignedTxn, error)
	// get txn content of txn-hash from p2p network
	SyncTxns([]Hash) error
	// broadcast txns to p2p network
	BroadcastTxns() error
	// remove txns after execute all tripods
	Remove() error
}

type ItxCache interface {
	Push(IsignedTxn) error
	Pop() (IsignedTxn, error)
}
