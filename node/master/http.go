package master

import (
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	. "yu/common"
	. "yu/node"
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

	var ip string
	if m.RunMode == MasterWorker {
		ip, err = m.findWorkerIP(tripodName, callName, ExecCall)
		if err != nil {
			c.String(
				http.StatusBadRequest,
				BadReqErrStr(tripodName, callName, err),
			)
			return
		}
	}

	err = m.txPool.Insert(ip, stxn)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	m.readyBcTxnsChan <- stxn
	c.String(http.StatusOK, "")
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

	var qerr error
	if m.RunMode == MasterWorker {
		var ip string
		ip, qerr = m.findWorkerIP(qcall.TripodName, qcall.QueryName, QryCall)
		forwardQueryToWorker(ip, c.Writer, c.Request)
	} else {
		qerr = m.land.Query(qcall)
	}
	if qerr != nil {
		c.String(
			http.StatusBadRequest,
			BadReqErrStr(qcall.TripodName, qcall.QueryName, qerr),
		)
		return
	}
}

func readPostBody(body io.ReadCloser) (JsonString, error) {
	byt, err := ioutil.ReadAll(body)
	return JsonString(byt), err
}
