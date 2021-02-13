package txpool

import . "yu/txn"

type ItxPool interface {
	Insert(tx IsignedTxn) error
	Package(numLimit uint64) ([]IsignedTxn, error)
	Remove() error
}
