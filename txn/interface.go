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

type SignedTxns []IsignedTxn

func (sts SignedTxns) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(sts)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (sts SignedTxns) Decode(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&sts)
	if err != nil {
		return err
	}
	return nil
}

type IsignedTxn interface {
	Size() int
	GetRaw() IunsignedTxn
	GetTxnHash() Hash
	GetPubkey() PubKey
	GetSignature() []byte
	Encode() ([]byte, error)
}
