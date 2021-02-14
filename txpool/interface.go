package txpool

import . "yu/txn"

type ItxPool interface {
	// insert into txCache for pending
	Pend(IsignedTxn) error
	// insert into txPool for tripods
	Insert(IsignedTxn) error
	// package some txns to send to tripods
	Package(numLimit uint64) ([]IsignedTxn, error)
	// pop pending txns
	Pop() (IsignedTxn, error)
	// remove txns after execute all tripods
	Remove() error
}

type ItxCache interface {
	Push(IsignedTxn) error
	Pop() (IsignedTxn, error)
}
