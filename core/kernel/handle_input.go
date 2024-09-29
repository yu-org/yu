package kernel

import (
	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/protocol"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/metrics"
)

// HandleTxn handles txn from outside.
// You can also self-define your input by calling HandleTxn (not only by default http and ws)
func (k *Kernel) HandleTxn(signedWrCall *protocol.SignedWrCall) error {
	stxn, err := NewSignedTxn(signedWrCall.Call, signedWrCall.Pubkey, signedWrCall.Address, signedWrCall.Signature)
	if err != nil {
		return err
	}
	wrCall := signedWrCall.Call
	_, err = k.Land.GetWriting(wrCall.TripodName, wrCall.FuncName)
	if err != nil {
		return err
	}

	err = k.handleTxnLocally(stxn)
	if err != nil {
		return err
	}

	go func() {
		err = k.pubUnpackedTxns(FromArray(stxn))
		if err != nil {
			logrus.Error("publish unpacked txns error: ", err)
		}
	}()

	return nil
}

func (k *Kernel) handleTxnLocally(stxn *SignedTxn) error {
	metrics.KernelHandleTxnCounter.WithLabelValues().Inc()
	k.mutex.Lock()
	defer k.mutex.Unlock()
	tri := k.Land.GetTripod(stxn.TripodName())
	if tri != nil {
		err := tri.PreTxnHandler.PreHandleTxn(stxn)
		if err != nil {
			return err
		}
	}
	if k.CheckReplayAttack(stxn) {
		return yerror.TxnDuplicated
	}
	err := k.Pool.CheckTxn(stxn)
	if err != nil {
		return err
	}
	return k.Pool.Insert(stxn)
}

func (k *Kernel) HandleRead(rdCall *common.RdCall) (*context.ResponseData, error) {
	ctx, err := context.NewReadContext(rdCall)
	if err != nil {
		return nil, err
	}

	rd, err := k.Land.GetReading(rdCall.TripodName, rdCall.FuncName)
	if err != nil {
		return nil, err
	}
	rd(ctx)
	return ctx.Response(), nil
}

func (k *Kernel) CheckReplayAttack(txn *SignedTxn) bool {
	if k.Pool.Exist(txn.TxnHash) {
		return true
	}
	if k.Chain.ChainID() != txn.ChainID() {
		return true
	}
	return k.TxDB.ExistTxn(txn.TxnHash)
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
