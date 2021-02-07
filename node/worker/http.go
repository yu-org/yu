package worker

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/common"
	. "yu/node"
)

func (w *Worker) HandleHttp() {
	r := gin.Default()

	r.GET(HeartbeatPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})

	// GET request
	r.GET(ExecApiPath, func(c *gin.Context) {
		w.DoHttpExecCall(c)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		w.DoHttpQryCall(c)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		w.DoHttpExecCall(c)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		w.DoHttpQryCall(c)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		w.DoHttpExecCall(c)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		w.DoHttpQryCall(c)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		w.DoHttpExecCall(c)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		w.DoHttpQryCall(c)
	})

	r.Run(w.httpPort)
}

func (w *Worker) DoHttpExecCall(c *gin.Context) {
	tripodName, execName := GetTripodCallName(c.Request)
	var ecallParams EcallParams
	err := c.ShouldBindJSON(&ecallParams)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("decode Execution Call error: %s", err.Error()),
		)
		return
	}
	ecall := &Ecall{
		TripodName: tripodName,
		ExecName:   execName,
		Params:     ecallParams,
	}
	err = w.land.Execute(ecall)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("Execution error: %s", err.Error()),
		)
		return
	}
	c.String(http.StatusOK, "")
}

func (w *Worker) DoHttpQryCall(c *gin.Context) {
	tripodName, qryName := GetTripodCallName(c.Request)
	var qcallParams QcallParams
	err := c.ShouldBindJSON(&qcallParams)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("decode Query Call error: %s", err.Error()),
		)
		return
	}
	qcall := &Qcall{
		TripodName: tripodName,
		QueryName:  qryName,
		Params:     qcallParams,
	}
	err = w.land.Query(qcall)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("Query error: %s", err.Error()),
		)
		return
	}
	c.String(http.StatusOK, "")
}
