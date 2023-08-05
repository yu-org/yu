package core

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	"net/http"
	"path/filepath"
	"strings"
)

// A complete writing-call url is POST /api/writing/{tripod}/{writing_name}
// A complete reading-call url is GET /api/reading/{tripod}/{reading_name}?xx=yy

const (
	// RootApiPath For developers, every customized Writing and Read of tripods
	// will base on '/api'.
	RootApiPath = "/api"

	BlockHashKey = "block_hash"
	PubkeyKey    = "pubkey"
	SignatureKey = "signature"
	LeiPriceKey  = "lei_price"
	TipsKey      = "tips"
)

var (
	WrApiPath      = filepath.Join(RootApiPath, WrCallType)
	RdApiPath      = filepath.Join(RootApiPath, RdCallType)
	SubResultsPath = "/subscribe/results"
)

// GetTripodCallName return (Tripod Name, Write/Read Name, error)
func GetTripodCallName(req *http.Request) (string, string, error) {
	path := req.URL.Path
	paths := strings.Split(path, "/")
	if len(paths) < 5 {
		return "", "", errors.New("URL path illegal")
	}
	return paths[3], paths[4], nil
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
