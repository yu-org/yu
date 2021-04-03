package common

import (
	"encoding/binary"
	"strconv"
	"unsafe"
	"yu/context"
)

type RunMode int

const (
	LocalNode RunMode = iota
	MasterWorker
)

const (
	StartBlockStage    = "Start Block"
	ExecuteTxnsStage   = "Execute Txns"
	EndBlockStage      = "End Block"
	FinalizeBlockStage = "Finalize Block"
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
	// todo: Query needs response
	Query func(*context.Context, Hash) error

	JsonString = string

	// The Call from clients, it is an instance of an 'Execution'.
	Ecall struct {
		TripodName string
		ExecName   string
		Params     JsonString
	}

	// The Call from clients, it is an instance of an 'Query'.
	Qcall struct {
		TripodName string
		QueryName  string
		BlockHash  Hash
		Params     JsonString
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

func BytesToBlockNum(byt []byte) BlockNum {
	u := binary.BigEndian.Uint64(byt)
	return BlockNum(u)
}

func StrToBlockNum(str string) (BlockNum, error) {
	bn, err := strconv.ParseUint(str, 10, 64)
	return BlockNum(bn), err
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
	bn = BytesToBlockNum(byt[:bnLen])
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
