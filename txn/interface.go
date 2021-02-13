package txn

import (
	. "yu/common"
)

type IunsignedTxn interface {
	Caller() Address
	Ecall() *Ecall
	Timestamp() int64
	Hash() (Hash, error)
	Encode() ([]byte, error)
}

type IsignedTxn interface {
}
