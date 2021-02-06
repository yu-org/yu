package master

import (
	"github.com/gin-gonic/gin"
	. "yu/common"
	. "yu/node"
)

func (m *Master) HandleHttp() {
	r := gin.Default()

	r.POST(RegisterNodeKeepersPath, func(c *gin.Context) {
		m.registerNodeKeepers(c)
	})

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

	r.Run(m.httpPort)
}
