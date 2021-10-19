package master

import (
	"github.com/gin-gonic/gin"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/node"
	"github.com/yu-org/yu/types"
	"net/http"
	"net/http/httputil"
)

func getQryInfoFromReq(req *http.Request, params string) (qcall *Qcall, err error) {
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

func getExecInfoFromReq(req *http.Request, params string) (tripodName, execName string, stxn *types.SignedTxn, err error) {
	tripodName, execName = GetTripodCallName(req)
	ecall := &Ecall{
		TripodName: tripodName,
		ExecName:   execName,
		Params:     params,
	}
	caller := GetAddress(req)
	sig := GetSignature(req)
	pubkey, err := GetPubkey(req)
	if err != nil {
		return
	}
	stxn, err = types.NewSignedTxn(caller, ecall, pubkey, sig)
	return
}

func getHttpJsonParams(c *gin.Context) (params string, err error) {
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
