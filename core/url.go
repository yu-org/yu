package core

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	. "github.com/yu-org/yu/common"
	"path/filepath"
)

// A complete writing-call url is POST /api/writing
// A complete reading-call url is POST /api/reading

const (
	// RootApiPath For developers, every customized Writing and Read of tripods
	// will base on '/api'.
	RootApiPath = "/api"
	WrCallType  = "writing"
	RdCallType  = "reading"

	TripodNameKey = "tripod_name"
	FuncNameKey   = "func_name"
	BlockHashKey  = "block_hash"
)

var (
	WrApiPath      = filepath.Join(RootApiPath, WrCallType)
	RdApiPath      = filepath.Join(RootApiPath, RdCallType)
	SubResultsPath = "/subscribe/results"
)

type SignedWrCall struct {
	Pubkey    []byte  `json:"pubkey"`
	Signature []byte  `json:"signature"`
	Call      *WrCall `json:"call"`
}

type WritingPostBody struct {
	// hex string
	Pubkey string `json:"pubkey"`
	// hex string
	Signature string  `json:"signature"`
	Call      *WrCall `json:"call"`
}

func GetSignedWrCall(ctx *gin.Context) (*SignedWrCall, error) {
	wpb := new(WritingPostBody)
	err := ctx.ShouldBindJSON(wpb)
	if err != nil {
		return nil, err
	}

	var pubkey []byte
	if wpb.Pubkey != "" {
		pubkey, err = hexutil.Decode(wpb.Pubkey)
		if err != nil {
			return nil, err
		}
	}

	var sig []byte
	if wpb.Signature != "" {
		sig, err = hexutil.Decode(wpb.Signature)
		if err != nil {
			return nil, err
		}
	}

	return &SignedWrCall{
		Pubkey:    pubkey,
		Signature: sig,
		Call:      wpb.Call,
	}, err
}

func GetRdCall(ctx *gin.Context) (*RdCall, error) {
	rdCall := new(RdCall)
	err := ctx.ShouldBindJSON(rdCall)
	return rdCall, err
}
