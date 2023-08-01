package kernel

import (
	"fmt"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	. "github.com/yu-org/yu/core/types"
	"net/http"
)

// HandleTxn handles txn from outside.
// You can also self-define your input by calling HandleTxn (not only by default http and ws)
func (k *Kernel) HandleTxn(stxn *SignedTxn) error {
	_, err := k.land.GetWriting(stxn.Raw.WrCall)
	if err != nil {
		return err
	}

	if k.txPool.Exist(stxn) {
		return err
	}

	err = k.txPool.CheckTxn(stxn)
	if err != nil {
		return err
	}

	go func() {
		err = k.pubUnpackedTxns(FromArray(stxn))
		if err != nil {
			logrus.Error("publish unpacked txns error: ", err)
		}
	}()

	return k.txPool.Insert(stxn)
}

func getRdFromHttp(req *http.Request, params string) (qcall *RdCall, err error) {
	tripodName, rdName := GetTripodCallName(req)
	blockHash := GetBlockHash(req)
	qcall = &RdCall{
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
	sig := GetSignature(req)
	pubkey, err := GetPubkey(req)
	if err != nil {
		return
	}
	stxn, err = NewSignedTxn(wrCall, pubkey, sig)
	return
}

func FindNoCallStr(tripodName, callName string, err error) string {
	return fmt.Sprintf("find Tripod(%s) Call(%s) error: %s", tripodName, callName, err.Error())
}
