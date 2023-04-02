package core

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	"net/http"
	"path/filepath"
)

const (
	// RootApiPath For developers, every customized Execution and Read of tripods
	// will base on '/api'.
	RootApiPath = "/api"

	TripodNameKey = "tripod"
	CallNameKey   = "call_name"
	AddressKey    = "address"
	BlockHashKey  = "block_hash"
	PubkeyKey     = "pubkey"
	SignatureKey  = "signature"
	LeiPriceKey   = "lei_price"
	TipsKey       = "tips"
)

var (
	WrApiPath      = filepath.Join(RootApiPath, WrCallType)
	RdApiPath      = filepath.Join(RootApiPath, RdCallType)
	SubResultsPath = "/subscribe/results"
)

// GetTripodCallName return (Tripod Name, Write/Read Name)
func GetTripodCallName(req *http.Request) (string, string) {
	query := req.URL.Query()
	return query.Get(TripodNameKey), query.Get(CallNameKey)
}

// GetAddress return the Address of Txn-Sender
func GetAddress(req *http.Request) Address {
	return HexToAddress(req.URL.Query().Get(AddressKey))
}

func GetBlockHash(req *http.Request) Hash {
	return HexToHash(req.URL.Query().Get(BlockHashKey))
}

func GetPubkey(req *http.Request) (keypair.PubKey, error) {
	pubkeyStr := req.URL.Query().Get(PubkeyKey)
	return keypair.PubkeyFromStr(pubkeyStr)
}

func GetSignature(req *http.Request) []byte {
	signStr := req.URL.Query().Get(SignatureKey)
	return FromHex(signStr)
}

func GetLeiPrice(req *http.Request) (uint64, error) {
	return hexutil.DecodeUint64(req.URL.Query().Get(LeiPriceKey))
}

func GetTips(req *http.Request) (uint64, error) {
	return hexutil.DecodeUint64(req.URL.Query().Get(TipsKey))
}
