package txn

import (
	"crypto/sha256"
	"time"
	. "yu/common"
	. "yu/utils/codec"
)

type UnsignedTxn struct {
	id        Hash
	caller    Address
	ecall     *Ecall
	timestamp int64
}

func NewUnsignedTxn(caller Address, ecall *Ecall) (IunsignedTxn, error) {
	UnsignedTxn := &UnsignedTxn{
		caller:    caller,
		ecall:     ecall,
		timestamp: time.Now().UnixNano(),
	}
	id, err := UnsignedTxn.Hash()
	if err != nil {
		return nil, err
	}
	UnsignedTxn.id = id
	return UnsignedTxn, nil
}

func (ut *UnsignedTxn) ID() Hash {
	return ut.id
}

func (ut *UnsignedTxn) Caller() Address {
	return ut.caller
}

func (ut *UnsignedTxn) Ecall() *Ecall {
	return ut.ecall
}

func (ut *UnsignedTxn) ToSignedTxn() (IsignedTxn, error) {

}

func (ut *UnsignedTxn) Timestamp() int64 {
	return ut.timestamp
}

func (ut *UnsignedTxn) Hash() (Hash, error) {
	var hash Hash
	byt, err := ut.Encode()
	if err != nil {
		return NullHash, err
	}
	hash = sha256.Sum256(byt)
	return hash, nil
}

func (ut *UnsignedTxn) Encode() ([]byte, error) {
	return GobEncode(ut)
}
