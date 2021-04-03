package handle

//import (
//	"net/http"
//	. "yu/common"
//	. "yu/node"
//	"yu/tripod"
//	"yu/txn"
//	. "yu/txpool"
//)
//
//func putTxpool(req *http.Request, params JsonString, txpool ItxPool, broadcastChan chan<- txn.IsignedTxn) error {
//	tripodName, execName := GetTripodCallName(req)
//	ecall := &Ecall{
//		TripodName: tripodName,
//		ExecName:   execName,
//		Params:     params,
//	}
//	caller := GetAddress(req)
//	pubkey, sig, err := GetPubkeyAndSignature(req)
//	if err != nil {
//		return err
//	}
//	stxn, err := txn.NewSignedTxn(caller, ecall, pubkey, sig)
//	if err != nil {
//		return err
//	}
//	err = txpool.Insert(stxn)
//	if err != nil {
//		return err
//	}
//	if broadcastChan != nil {
//		broadcastChan <- stxn
//	}
//
//	return nil
//}
//
//func doQryCall(req *http.Request, params JsonString, land *tripod.Land) error {
//	tripodName, qryName := GetTripodCallName(req)
//	blockNum, err := GetBlockNumber(req)
//	if err != nil {
//		return err
//	}
//	qcall := &Qcall{
//		TripodName:  tripodName,
//		QueryName:   qryName,
//		Params:      params,
//		BlockNumber: blockNum,
//	}
//	return land.Query(qcall)
//}
