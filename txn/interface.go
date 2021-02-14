package txn

import (
	. "yu/common"
	. "yu/keypair"
)

type IunsignedTxn interface {
	Caller() Address
	Ecall() *Ecall
	Timestamp() int64
	Hash() (Hash, error)
	ToSignedTxn() (IsignedTxn, error)
	Encode() ([]byte, error)
}

type IsignedTxn interface {
	GetRaw() IunsignedTxn
	GetTxnHash() Hash
	GetPubkey() PubKey
	GetSignature() []byte
}
