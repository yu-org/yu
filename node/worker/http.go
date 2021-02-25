package worker

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/node"
	. "yu/node/handle"
)

func (w *Worker) HandleHttp() {
	r := gin.Default()

	r.GET(HeartbeatPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})

	// GET request
	r.GET(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, w.txPool)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, w.land)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, w.txPool)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, w.land)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, w.txPool)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, w.land)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, w.txPool)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, w.land)
	})

	r.Run(w.httpPort)
}
