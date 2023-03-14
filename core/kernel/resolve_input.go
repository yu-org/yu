package kernel

import (
	"fmt"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	"github.com/yu-org/yu/core/types"
	"net/http"
)

type (
	ResolveReading func(input any, a ...any) (*Rdcall, error)
	ResolveWriting func(input any, a ...any) (*types.SignedTxn, error)
)

func getRdInfoFromReq(req *http.Request, params string) (qcall *Rdcall, err error) {
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

func getWrInfoFromReq(req *http.Request, params string) (stxn *types.SignedTxn, err error) {
	tripodName, wrName := GetTripodCallName(req)
	leiPrice, err := GetLeiPrice(req)
	if err != nil {
		return
	}
	wrCall := &WrCall{
		TripodName:  tripodName,
		WritingName: wrName,
		Params:      params,
		LeiPrice:    leiPrice,
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
