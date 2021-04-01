package master

import (
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	. "yu/common"
	. "yu/node"
	. "yu/txn"
	. "yu/utils/error_handle"
)

const PARAMS_KEY = "params"

func (m *Master) HandleHttp() {
	r := gin.Default()

	r.POST(RegisterNodeKeepersPath, func(c *gin.Context) {
		m.registerNodeKeepers(c)
	})

	// GET request
	r.GET(ExecApiPath, func(c *gin.Context) {
		m.handleHttpExec(c)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		m.handleHttpQry(c)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		m.handleHttpExec(c)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		m.handleHttpQry(c)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		m.handleHttpExec(c)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		m.handleHttpQry(c)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		m.handleHttpExec(c)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		m.handleHttpQry(c)
	})

	r.Run(m.httpPort)
}

func (m *Master) handleHttpExec(c *gin.Context) {
	params, err := getHttpJsonParams(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tripodName, callName, stxn, err := getExecInfoFromReq(c.Request, params)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var (
		ip   string
		name string
	)
	if m.RunMode == MasterWorker {
		ip, name, err = m.findWorkerIpAndName(tripodName, callName, ExecCall)
		if err != nil {
			c.String(
				http.StatusBadRequest,
				BadReqErrStr(tripodName, callName, err),
			)
			return
		}
	}

	err = m.txPool.Insert(name, stxn)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	fmap := make(map[string]*TxnsAndWorkerName)
	fmap[ip] = &TxnsAndWorkerName{
		Txns:       []*SignedTxn{stxn},
		WorkerName: name,
	}
	err = m.forwardTxnsForCheck(fmap)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	m.readyBcTxnsChan <- stxn
}

func (m *Master) handleHttpQry(c *gin.Context) {
	params, err := getHttpJsonParams(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	qcall, err := getQryInfoFromReq(c.Request, params)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if m.RunMode == MasterWorker {
		var ip string
		ip, err = m.findWorkerIP(qcall.TripodName, qcall.QueryName, QryCall)
		if err != nil {
			c.String(
				http.StatusBadRequest,
				BadReqErrStr(qcall.TripodName, qcall.QueryName, err),
			)
			return
		}
		forwardQueryToWorker(ip, c.Writer, c.Request)
	} else {
		err = m.land.Query(qcall)
		if err != nil {
			c.String(
				http.StatusBadRequest,
				BadReqErrStr(qcall.TripodName, qcall.QueryName, err),
			)
			return
		}
	}

}

func readPostBody(body io.ReadCloser) (JsonString, error) {
	byt, err := ioutil.ReadAll(body)
	return JsonString(byt), err
}
