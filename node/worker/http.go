package worker

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/node"
)

func (w *Worker) HandleHttp() {
	r := gin.Default()

	r.GET(HeartbeatPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})

	// GET request
	r.GET(ExecApiHttpPath, func(c *gin.Context) {
		tripodName, execName := ResolveHttpApiUrl(c)
	})
	r.GET(QryApiHttpPath, func(c *gin.Context) {
		tripodName, qryName := ResolveHttpApiUrl(c)
	})

	// POST request
	r.POST(ExecApiHttpPath, func(c *gin.Context) {
		tripodName, execName := ResolveHttpApiUrl(c)
	})
	r.POST(QryApiHttpPath, func(c *gin.Context) {
		tripodName, qryName := ResolveHttpApiUrl(c)
	})

	// PUT request
	r.PUT(ExecApiHttpPath, func(c *gin.Context) {
		tripodName, execName := ResolveHttpApiUrl(c)
	})
	r.PUT(QryApiHttpPath, func(c *gin.Context) {
		tripodName, qryName := ResolveHttpApiUrl(c)
	})

	// DELETE request
	r.DELETE(ExecApiHttpPath, func(c *gin.Context) {
		tripodName, execName := ResolveHttpApiUrl(c)
	})
	r.DELETE(QryApiHttpPath, func(c *gin.Context) {
		tripodName, qryName := ResolveHttpApiUrl(c)
	})

	r.Run(w.HttpPort)
}
