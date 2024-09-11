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
	FullNode int = iota
	LightNode
	ArchiveNode
)

const (
	StartBlockStage    = "Start Block"
	ExecuteTxnsStage   = "Execute Txns"
	EndBlockStage      = "End Block"
	FinalizeBlockStage = "Finalize Block"
)

type (
	BlockNum uint32
	// BlockId Uses to be a Key to store into KVDB.
	// Add BlockHash to the BlockNum's end.
	BlockId [BlockIdLen]byte

	// WrCall from clients, it is an instance of an 'Writing'.
	WrCall struct {
		ChainID    uint64 `json:"chain_id"`
		TripodName string `json:"tripod_name"`
		FuncName   string `json:"func_name"`
		Params     string `json:"params"`
		// TODO: make LeiPrice and Tips as a sortable interface.
		LeiPrice uint64 `json:"lei_price,omitempty"`
		Tips     uint64 `json:"tips,omitempty"`
	}

	// RdCall from clients, it is an instance of an 'Read'.
	RdCall struct {
		TripodName string `json:"tripod_name"`
		FuncName   string `json:"func_name"`
		Params     string `json:"params"`
		BlockHash  string `json:"block_hash,omitempty"`
	}
	// CallType is Writing or Reading
	CallType int
)

func (e *WrCall) BindJsonParams(v interface{}) error {
	return BindJsonParams(e.Params, v)
}

func (e *WrCall) Hash() ([]byte, error) {
	byt, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(byt)
	return hash[:], nil
}

func BindJsonParams(params string, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader([]byte(params)))
	d.UseNumber()
	return d.Decode(v)
}

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

const Separator = "|"

// HashesToHex returns hash_hex|hash_hex|...
func HashesToHex(hs []Hash) string {
	var buffer bytes.Buffer
	for _, h := range hs {
		buffer.WriteString(ToHex(h.Bytes()))
		buffer.WriteString(Separator)
	}

	hexHashes := buffer.String()
	return strings.TrimSuffix(hexHashes, Separator)
}

func HexToHashes(s string) (hs []Hash) {
	arr := strings.Split(s, Separator)
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
