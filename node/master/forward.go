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
	tripodName, callName := GetTripodCallName(c.Request)
	ip, err := m.findWorkerIP(tripodName, callName, callType)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			badReqErrStr(tripodName, callName, err),
		)
		return
	}
	m.forwardToWorker(ip, c.Writer, c.Request)
	c.String(http.StatusOK, "")
}

func (m *Master) forwardWsCall(w http.ResponseWriter, req *http.Request, callType CallType) {
	tripodName, callName := GetTripodCallName(req)
	ip, err := m.findWorkerIP(tripodName, callName, callType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(badReqErrStr(tripodName, callName, err)))
		return
	}
	m.forwardToWorker(ip, w, req)
	w.WriteHeader(http.StatusOK)
}

func (m *Master) forwardToWorker(ip string, rw http.ResponseWriter, req *http.Request) {
	director := func(req *http.Request) {
		req.URL.Host = ip
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(rw, req)
}

func badReqErrStr(tripodName, callName string, err error) string {
	return fmt.Sprintf("find Tripod(%s) Call(%s) error: %s", tripodName, callName, err.Error())
}
