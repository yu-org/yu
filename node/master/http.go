package master

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"yu/common"
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
		m.forwardCall(c, common.ExecCall)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		m.forwardCall(c, common.QryCall)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		m.forwardCall(c, common.ExecCall)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		m.forwardCall(c, common.QryCall)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		m.forwardCall(c, common.ExecCall)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		m.forwardCall(c, common.QryCall)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		m.forwardCall(c, common.ExecCall)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		m.forwardCall(c, common.QryCall)
	})

	r.Run(m.port)
}

func (m *Master) forwardCall(c *gin.Context, callType CallType) {
	tripodName, callName := ResolveApiUrl(c)
	ip, err := m.findWorkerIP(tripodName, callName, callType)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("find Tripod(%s) Call(%s) error: %s", tripodName, callName, err.Error()),
		)
		return
	}
	m.forwardToWorker(ip, c)
	c.String(http.StatusOK, "")
}

func (m *Master) forwardToWorker(ip string, c *gin.Context) {
	director := func(req *http.Request) {
		req.URL.Host = ip
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}
