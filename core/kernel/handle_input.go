package kernel

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core"
	"github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/types"
)

// HandleTxn handles txn from outside.
// You can also self-define your input by calling HandleTxn (not only by default http and ws)
func (k *Kernel) HandleTxn(signedWrCall *core.SignedWrCall) error {
	stxn, err := NewSignedTxn(signedWrCall.Call, signedWrCall.Pubkey, signedWrCall.Signature)
	if err != nil {
		return err
	}
	wrCall := signedWrCall.Call
	_, err = k.land.GetWriting(wrCall.TripodName, wrCall.FuncName)
	if err != nil {
		return err
	}

	if k.Pool.Exist(stxn) {
		return nil
	}

	err = k.Pool.CheckTxn(stxn)
	if err != nil {
		return err
	}

	go func() {
		err = k.pubUnpackedTxns(FromArray(stxn))
		if err != nil {
			logrus.Error("publish unpacked txns error: ", err)
		}
	}()

	return k.Pool.Insert(stxn)
}

func (k *Kernel) HandleRead(rdCall *common.RdCall) (*context.ResponseData, error) {
	ctx, err := context.NewReadContext(rdCall)
	if err != nil {
		return nil, err
	}

	rd, err := k.land.GetReading(rdCall.TripodName, rdCall.FuncName)
	if err != nil {
		return nil, err
	}
	rd(ctx)
	return ctx.Response(), nil
}

//func getRdFromHttp(req *http.Request, params string) (rdCall *RdCall, err error) {
//	tripodName, rdName, urlErr := GetTripodCallName(req)
//	if err != nil {
//		return nil, urlErr
//	}
//	blockHash := GetBlockHash(req)
//	rdCall = &RdCall{
//		TripodName:  tripodName,
//		FuncName: rdName,
//		Params:      params,
//		BlockHash:   blockHash,
//	}
//	return
//}

//func getWrFromHttp(req *http.Request, params string) (stxn *SignedTxn, err error) {
//	tripodName, wrName, urlErr := GetTripodCallName(req)
//	if err != nil {
//		return nil, urlErr
//	}
//	leiPrice, err := GetLeiPrice(req)
//	if err != nil {
//		return
//	}
//	tips, err := GetTips(req)
//	wrCall := &WrCall{
//		TripodName: tripodName,
//		FuncName:   wrName,
//		Params:     params,
//		LeiPrice:   leiPrice,
//		Tips:       tips,
//	}
//	sig := GetSignature(req)
//	pubkey, err := GetPubkey(req)
//	if err != nil {
//		return
//	}
//	stxn, err = NewSignedTxn(wrCall, pubkey, sig)
//	return
//}
