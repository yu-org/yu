package kernel

import (
	"fmt"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	. "github.com/yu-org/yu/core/types"
	"net/http"
)

type (
	ResolveRd  func(input any, a ...any) (*Rdcall, error)
	ResolveTxn func(input any, a ...any) (*SignedTxn, error)
)

var (
	RdResolves  = make([]ResolveRd, 0)
	TxnResolves = make([]ResolveTxn, 0)
)

func SetRdResolves(rds ...ResolveRd) {
	RdResolves = append(RdResolves, rds...)
}

func SetTxnResolves(wrs ...ResolveTxn) {
	TxnResolves = append(TxnResolves, wrs...)
}

func getRdFromHttp(req *http.Request, params string) (qcall *Rdcall, err error) {
	tripodName, rdName := GetTripodCallName(req)
	blockHash := GetBlockHash(req)
	qcall = &Rdcall{
		TripodName:  tripodName,
		ReadingName: rdName,
		Params:      params,
		BlockHash:   blockHash,
	}
	return
}

func getWrFromHttp(req *http.Request, params string) (stxn *SignedTxn, err error) {
	tripodName, wrName := GetTripodCallName(req)
	leiPrice, err := GetLeiPrice(req)
	if err != nil {
		return
	}
	tips, err := GetTips(req)
	wrCall := &WrCall{
		TripodName:  tripodName,
		WritingName: wrName,
		Params:      params,
		LeiPrice:    leiPrice,
		Tips:        tips,
	}
	caller := GetAddress(req)
	sig := GetSignature(req)
	pubkey, err := GetPubkey(req)
	if err != nil {
		return
	}
	stxn, err = NewSignedTxn(caller, wrCall, pubkey, sig)
	return
}

func FindNoCallStr(tripodName, callName string, err error) string {
	return fmt.Sprintf("find Tripod(%s) Call(%s) error: %s", tripodName, callName, err.Error())
}
