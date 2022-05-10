package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"strings"
	"unsafe"
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
	BlockNum uint32
	// Use to be a Key to store into KVDB.
	// Add BlockHash to the BlockNum's end.
	BlockId [BlockIdLen]byte

	// The Call from clients, it is an instance of an 'Execution'.
	Ecall struct {
		TripodName string
		ExecName   string
		Params     string
		// TODO: make LeiPrice as a sortable interface.
		LeiPrice uint64
	}

	// The Call from clients, it is an instance of an 'Query'.
	Qcall struct {
		TripodName string
		QueryName  string
		BlockHash  Hash
		Params     string
	}
	// Execution or Query
	CallType int
)

func (e *Ecall) Hash() ([]byte, error) {
	byt, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(byt)
	return hash[:], nil
}

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
	binary.BigEndian.PutUint32(byt, uint32(bn))
	return byt
}

func BytesToBlockNum(byt []byte) BlockNum {
	u := binary.BigEndian.Uint32(byt)
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

var (
	NullHash    Hash    = [HashLen]byte{}
	NullAddress Address = [AddressLen]byte{}
)

func HashesToHex(hs []Hash) string {
	var buffer bytes.Buffer
	for _, h := range hs {
		buffer.WriteString(ToHex(h.Bytes()))
	}
	return buffer.String()
}

func HexToHashes(s string) (hs []Hash) {
	arr := strings.SplitN(s, "", HashLen)
	for _, hx := range arr {
		hs = append(hs, HexToHash(hx))
	}
	return
}

func HashesToBytes(hs []Hash) []byte {
	return []byte(HashesToHex(hs))
}

func BytesToHashes(data []byte) []Hash {
	return HexToHashes(string(data))
}

func HashesToTwoBytes(hs []Hash) (byts [][]byte) {
	for _, h := range hs {
		byts = append(byts, h.Bytes())
	}
	return
}

func TwoBytesToHashes(byts [][]byte) (hs []Hash) {
	for _, byt := range byts {
		hs = append(hs, BytesToHash(byt))
	}
	return
}
