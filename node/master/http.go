package master

import (
	"github.com/gin-gonic/gin"
	. "yu/common"
	. "yu/node"
	. "yu/node/handle"
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
		PutHttpInTxpool(c, m.txPool)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, m.land)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, m.txPool)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, m.land)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, m.txPool)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, m.land)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, m.txPool)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, m.land)
	})
}

func (m *Master) handleByMasterWorker(r *gin.Engine) {
	// GET request
	r.GET(ExecApiPath, func(c *gin.Context) {
		m.forwardHttpCall(c, ExecCall)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		m.forwardHttpCall(c, QryCall)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		m.forwardHttpCall(c, ExecCall)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		m.forwardHttpCall(c, QryCall)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		m.forwardHttpCall(c, ExecCall)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		m.forwardHttpCall(c, QryCall)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		m.forwardHttpCall(c, ExecCall)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		m.forwardHttpCall(c, QryCall)
	})
}
