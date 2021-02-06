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
	r.GET(ExecApiHttpPath, func(c *gin.Context) {
		m.forwardHttpCall(c, ExecCall)
	})
	r.GET(QryApiHttpPath, func(c *gin.Context) {
		m.forwardHttpCall(c, QryCall)
	})

	// POST request
	r.POST(ExecApiHttpPath, func(c *gin.Context) {
		m.forwardHttpCall(c, ExecCall)
	})
	r.POST(QryApiHttpPath, func(c *gin.Context) {
		m.forwardHttpCall(c, QryCall)
	})

	// PUT request
	r.PUT(ExecApiHttpPath, func(c *gin.Context) {
		m.forwardHttpCall(c, ExecCall)
	})
	r.PUT(QryApiHttpPath, func(c *gin.Context) {
		m.forwardHttpCall(c, QryCall)
	})

	// DELETE request
	r.DELETE(ExecApiHttpPath, func(c *gin.Context) {
		m.forwardHttpCall(c, ExecCall)
	})
	r.DELETE(QryApiHttpPath, func(c *gin.Context) {
		m.forwardHttpCall(c, QryCall)
	})

	r.Run(m.httpPort)
}
