package master

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	. "yu/common"
	. "yu/node"
)

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
