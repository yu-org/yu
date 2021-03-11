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

	//------------- requests from client ----------------

	// GET request
	r.GET(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, w.txPool, w.readyBcTxnsChan)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, w.land)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, w.txPool, w.readyBcTxnsChan)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, w.land)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, w.txPool, w.readyBcTxnsChan)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, w.land)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		PutHttpInTxpool(c, w.txPool, w.readyBcTxnsChan)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		DoHttpQryCall(c, w.land)
	})

	// ------------ Process Block ---------------
	r.GET(StartBlockPath, func(c *gin.Context) {

	})

	r.GET(EndBlockPath, func(c *gin.Context) {

	})

	r.GET(ExecuteTxnsPath, func(c *gin.Context) {

	})

	r.GET(FinalizeBlockPath, func(c *gin.Context) {

	})

	//------------- requests from P2P network ---------------

	// block from P2P
	r.POST(BlockFromP2P, func(c *gin.Context) {

	})

	// txns from P2P
	r.POST(TxnsFromP2P, func(c *gin.Context) {
		w.CheckTxnsFromP2P(c)
	})

	r.Run(w.httpPort)
}
