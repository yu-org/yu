package common

import (
	"encoding/binary"
	"unsafe"
	"yu/context"
)

type (
	BlockNum uint64
	// Use to be a Key to store into KVDB.
	// Add BlockHash to the BlockNum's end.
	BlockId [BlockIdLen]byte
	// Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(*context.Context) error
	// Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	Query func(*context.Context, BlockNum) error
	// The Call from clients, it is an instance of an 'Execution'.
	Ecall struct {
		TripodName string
		ExecName   string
		Params     []interface{}
	}
	// The Call from clients, it is an instance of an 'Query'.
	Qcall struct {
		TripodName  string
		QueryName   string
		Params      []interface{}
		BlockNumber BlockNum
	}
	// Execution or Query
	CallType int
)

const (
	ExecCall CallType = iota
	QryCall

	ExecCallType = "execution"
	QryCallType  = "query"
)

func (bn BlockNum) len() int {
	return int(unsafe.Sizeof(bn))
}

func (bn BlockNum) Bytes() []byte {
	byt := make([]byte, bn.len())
	binary.BigEndian.PutUint64(byt, uint64(bn))
	return byt
}

func ToBlockNum(byt []byte) BlockNum {
	u := binary.BigEndian.Uint64(byt)
	return BlockNum(u)
}

func NewBlockId(bn BlockNum, hash Hash) BlockId {
	bnByt := bn.Bytes()
	hashByt := hash.Bytes()

	var id BlockId
	bnLen := bn.len()
	copy(id[:bnLen], bnByt)
	copy(id[bnLen:], hashByt)
	return id
}

func DecodeBlockId(byt []byte) (id BlockId) {
	copy(id[:], byt)
	return
}

func (bi BlockId) Bytes() []byte {
	return bi[:]
}

func (bi BlockId) Separate() (bn BlockNum, hash Hash) {
	byt := bi.Bytes()
	bnLen := bn.len()
	bn = ToBlockNum(byt[:bnLen])
	copy(hash[:], byt[bnLen:])
	return
}

const (
	// HashLength is the expected length of the hash
	HashLen = 32
	// AddressLength is the expected length of the address
	AddressLen = 20
	// len(BlockNum) + HashLen = 40
	BlockIdLen = 40
)

type (
	Hash    [HashLen]byte
	Address [AddressLen]byte
)

var NullHash Hash = [HashLen]byte{}
