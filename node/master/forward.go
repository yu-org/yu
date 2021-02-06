package master

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	. "yu/common"
	. "yu/node"
)

func (m *Master) forwardHttpCall(c *gin.Context, callType CallType) {
	tripodName, callName := ResolveHttpApiUrl(c)
	ip, err := m.findWorkerIP(tripodName, callName, callType)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("find Tripod(%s) Call(%s) error: %s", tripodName, callName, err.Error()),
		)
		return
	}
	m.forwardToWorker(ip, c.Writer, c.Request)
	c.String(http.StatusOK, "")
}

func (m *Master) forwardWsCall(w http.ResponseWriter, req *http.Request, callType CallType) {
	tripodName, callName := ResolveWsApiUrl(req)
	ip, err := m.findWorkerIP(tripodName, callName, callType)
	if err != nil {

	}
	m.forwardToWorker(ip, w, req)

}

func (m *Master) forwardToWorker(ip string, rw http.ResponseWriter, req *http.Request) {
	director := func(req *http.Request) {
		req.URL.Host = ip
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(rw, req)
}
