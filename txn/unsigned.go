package txn

import (
	"crypto/sha256"
	"time"
	. "yu/common"
	. "yu/utils/codec"
)

type UnsignedTxn struct {
	Id        Hash
	Caller    Address
	Ecall     *Ecall
	Timestamp int64
}

func NewUnsignedTxn(caller Address, ecall *Ecall) (*UnsignedTxn, error) {
	UnsignedTxn := &UnsignedTxn{
		Caller:    caller,
		Ecall:     ecall,
		Timestamp: time.Now().UnixNano(),
	}
	id, err := UnsignedTxn.Hash()
	if err != nil {
		return nil, err
	}
	UnsignedTxn.Id = id
	return UnsignedTxn, nil
}

func (ut *UnsignedTxn) ID() Hash {
	return ut.Id
}

func (ut *UnsignedTxn) GetCaller() Address {
	return ut.Caller
}

func (ut *UnsignedTxn) GetEcall() *Ecall {
	return ut.Ecall
}

func (ut *UnsignedTxn) GetTimestamp() int64 {
	return ut.Timestamp
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
	return GlobalCodec.EncodeToBytes(ut)
}

func (ut *UnsignedTxn) Decode(data []byte) (*UnsignedTxn, error) {
	err := GlobalCodec.DecodeBytes(data, ut)
	return ut, err
}
