package master

import (
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/node"
	. "github.com/Lawliet-Chan/yu/txn"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
)

func getQryInfoFromReq(req *http.Request, params JsonString) (qcall *Qcall, err error) {
	tripodName, qryName := GetTripodCallName(req)
	blockHash := GetBlockHash(req)
	qcall = &Qcall{
		TripodName: tripodName,
		QueryName:  qryName,
		Params:     params,
		BlockHash:  blockHash,
	}
	return
}

func getExecInfoFromReq(req *http.Request, params JsonString) (tripodName, execName string, stxn *SignedTxn, err error) {
	tripodName, callName := GetTripodCallName(req)
	ecall := &Ecall{
		TripodName: tripodName,
		ExecName:   callName,
		Params:     params,
	}
	caller := GetAddress(req)
	sig := GetSignature(req)
	pubkey, err := GetPubkey(req)
	if err != nil {
		return
	}
	stxn, err = NewSignedTxn(caller, ecall, pubkey, sig)
	return
}

func getHttpJsonParams(c *gin.Context) (params JsonString, err error) {
	if c.Request.Method == http.MethodPost {
		params, err = readPostBody(c.Request.Body)
		if err != nil {
			return
		}
	} else {
		params = c.GetString(PARAMS_KEY)
	}
	return
}

func forwardQueryToWorker(ip string, rw http.ResponseWriter, req *http.Request) {
	director := func(req *http.Request) {
		req.URL.Host = ip
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(rw, req)
}
