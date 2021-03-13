package master

import (
	"github.com/gin-gonic/gin"
	"net/http"
	. "yu/common"
	. "yu/node"
	. "yu/node/handle"
	. "yu/utils/error_handle"
)

func (m *Master) HandleHttp() {
	r := gin.Default()

	r.POST(RegisterNodeKeepersPath, func(c *gin.Context) {
		m.registerNodeKeepers(c)
	})
	switch m.RunMode {
	case LocalNode:
		m.handleByLocal(r)
	case MasterWorker:
		m.handleByMasterWorker(r)
	}

	r.Run(m.httpPort)
}

func (m *Master) handleByLocal(r *gin.Engine) {
	// GET request
	r.GET(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, m.txPool, m.readyBcTxnsChan)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, m.land)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, m.txPool, m.readyBcTxnsChan)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, m.land)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, m.txPool, m.readyBcTxnsChan)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, m.land)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, m.txPool, m.readyBcTxnsChan)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, m.land)
	})
}

func (m *Master) handleByMasterWorker(r *gin.Engine) {
	// GET request
	r.GET(ExecApiPath, func(c *gin.Context) {
		m.handleHttpCall(c, ExecCall)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		m.handleHttpCall(c, QryCall)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		m.handleHttpCall(c, ExecCall)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		m.handleHttpCall(c, QryCall)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		m.handleHttpCall(c, ExecCall)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		m.handleHttpCall(c, QryCall)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		m.handleHttpCall(c, ExecCall)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		m.handleHttpCall(c, QryCall)
	})
}

func (m *Master) handleHttpCall(c *gin.Context, callType CallType) {
	tripodName, callName := GetTripodCallName(c.Request)
	ip, err := m.findWorkerIP(tripodName, callName, callType)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			BadReqErrStr(tripodName, callName, err),
		)
		return
	}
	if callType == ExecCall {
		PutHttpInTxpool(c, m.txPool, m.readyBcTxnsChan)
	}
	c.String(http.StatusOK, "")
}
