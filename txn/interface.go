package txn

import (
	"bytes"
	"encoding/gob"
	. "yu/common"
	. "yu/keypair"
)

type IunsignedTxn interface {
	ID() Hash
	Caller() Address
	Ecall() *Ecall
	Timestamp() int64
	Hash() (Hash, error)
	ToSignedTxn() (IsignedTxn, error)
	Encode() ([]byte, error)
	Decode(data []byte) error
}

type IsignedTxn interface {
	Size() int
	GetRaw() IunsignedTxn
	GetTxnHash() Hash
	GetPubkey() PubKey
	GetSignature() []byte
	Encode() ([]byte, error)
}
