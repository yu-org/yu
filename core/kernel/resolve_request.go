package kernel

import (
	"fmt"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	"github.com/yu-org/yu/core/types"
	"net/http"
)

func getQryInfoFromReq(req *http.Request, params string) (qcall *Rdcall, err error) {
	tripodName, qryName := GetTripodCallName(req)
	blockHash := GetBlockHash(req)
	qcall = &Rdcall{
		TripodName: tripodName,
		QueryName:  qryName,
		Params:     params,
		BlockHash:  blockHash,
	}
	return
}

func getExecInfoFromReq(req *http.Request, params string) (tripodName, execName string, stxn *types.SignedTxn, err error) {
	tripodName, execName = GetTripodCallName(req)
	leiPrice, err := GetLeiPrice(req)
	if err != nil {
		return
	}
	wrCall := &WrCall{
		TripodName: tripodName,
		ExecName:   execName,
		Params:     params,
		LeiPrice:   leiPrice,
	}
	caller := GetAddress(req)
	sig := GetSignature(req)
	pubkey, err := GetPubkey(req)
	if err != nil {
		return
	}
	stxn, err = types.NewSignedTxn(caller, wrCall, pubkey, sig)
	return
}

func FindNoCallStr(tripodName, callName string, err error) string {
	return fmt.Sprintf("find Tripod(%s) Call(%s) error: %s", tripodName, callName, err.Error())
}
