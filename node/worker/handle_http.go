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
	r.GET(ExecApiPath, func(c *gin.Context) {
		tripodName, execName := ResolveApiUrl(c)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		tripodName, qryName := ResolveApiUrl(c)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		tripodName, execName := ResolveApiUrl(c)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		tripodName, qryName := ResolveApiUrl(c)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		tripodName, execName := ResolveApiUrl(c)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		tripodName, qryName := ResolveApiUrl(c)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		tripodName, execName := ResolveApiUrl(c)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		tripodName, qryName := ResolveApiUrl(c)
	})

	r.Run(w.ServesPort)
}
