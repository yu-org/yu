package core

import (
	"github.com/gin-gonic/gin"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	"path/filepath"
)

// A complete writing-call url is POST /api/writing
// A complete reading-call url is GET /api/reading?tripod={tripod}&func_name={func_name}?xx=yy

const (
	// RootApiPath For developers, every customized Writing and Read of tripods
	// will base on '/api'.
	RootApiPath = "/api"
	WrCallType  = "writing"
	RdCallType  = "reading"

	TripodKey    = "tripod"
	FuncNameKey  = "func_name"
	BlockHashKey = "block_hash"
)

var (
	WrApiPath      = filepath.Join(RootApiPath, WrCallType)
	RdApiPath      = filepath.Join(RootApiPath, RdCallType)
	SubResultsPath = "/subscribe/results"
)

type RawWrCall struct {
	Pubkey    keypair.PubKey `json:"pubkey"`
	Signature []byte         `json:"signature"`
	Call      *WrCall        `json:"call"`
}

type WritingPostBody struct {
	// hex string
	Pubkey string `json:"pubkey"`
	// hex string
	Signature string  `json:"signature"`
	Call      *WrCall `json:"call"`
}

func GetRawWrCall(ctx *gin.Context) (*RawWrCall, error) {
	wpb := new(WritingPostBody)
	err := ctx.ShouldBindJSON(wpb)
	if err != nil {
		return nil, err
	}
	pubkey, err := keypair.PubkeyFromStr(wpb.Pubkey)
	if err != nil {
		return nil, err
	}
	return &RawWrCall{
		Pubkey:    pubkey,
		Signature: FromHex(wpb.Signature),
		Call:      wpb.Call,
	}, err
}

func GetRdCall(ctx *gin.Context) (*RdCall, error) {
	tri := ctx.GetString(TripodKey)
	fn := ctx.GetString(FuncNameKey)
	return &RdCall{
		TripodName: tri,
		FuncName:   fn,
	}, nil
}
