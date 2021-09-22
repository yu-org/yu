package txn

import (
	"crypto/sha256"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/utils/codec"
	ytime "github.com/yu-org/yu/utils/time"
)

type UnsignedTxn struct {
	Id        Hash
	Caller    Address
	Ecall     *Ecall
	Timestamp uint64
}

func NewUnsignedTxn(caller Address, ecall *Ecall) (*UnsignedTxn, error) {
	utxn := &UnsignedTxn{
		Caller:    caller,
		Ecall:     ecall,
		Timestamp: ytime.NowNanoTsU64(),
	}
	id, err := utxn.Hash()
	if err != nil {
		return nil, err
	}
	utxn.Id = id
	return utxn, nil
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

func (ut *UnsignedTxn) GetTimestamp() uint64 {
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
