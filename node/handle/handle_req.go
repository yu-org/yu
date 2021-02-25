package handle

import (
	"net/http"
	. "yu/common"
	. "yu/node"
	"yu/tripod"
	"yu/txn"
	. "yu/txpool"
)

func putTxpool(req *http.Request, params JsonString, txpool ItxPool) error {
	tripodName, execName := GetTripodCallName(req)
	ecall := &Ecall{
		TripodName: tripodName,
		ExecName:   execName,
		Params:     params,
	}
	caller := GetAddress(req)
	utxn, err := txn.NewUnsignedTxn(caller, ecall)
	if err != nil {
		return err
	}
	stxn, err := utxn.ToSignedTxn()
	if err != nil {
		return err
	}
	err = txpool.Insert(stxn)
	if err != nil {
		return err
	}
	txpool.BroadcastTxn(stxn)
	return nil
}

func doQryCall(req *http.Request, params JsonString, land *tripod.Land) error {
	tripodName, qryName := GetTripodCallName(req)
	blockNum, err := GetBlockNumber(req)
	if err != nil {
		return err
	}
	qcall := &Qcall{
		TripodName:  tripodName,
		QueryName:   qryName,
		Params:      params,
		BlockNumber: blockNum,
	}
	return land.Query(qcall)
}
