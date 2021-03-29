package worker

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	. "yu/blockchain"
	. "yu/node"
	. "yu/node/handle"
	"yu/tripod"
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
	r.POST(StartBlockPath, func(c *gin.Context) {
		block, err := DecodeBlockFromHttp(c.Request.Body, w.chain)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = w.land.RangeList(func(tri tripod.Tripod) error {
			return tri.StartBlock(w.chain, block, w.txPool)
		})
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		encodeBlockBackHttp(c, block)
	})

	r.POST(EndBlockPath, func(c *gin.Context) {
		block, err := DecodeBlockFromHttp(c.Request.Body, w.chain)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = w.land.RangeList(func(tri tripod.Tripod) error {
			return tri.EndBlock(w.chain, block)
		})
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		encodeBlockBackHttp(c, block)
	})

	r.POST(ExecuteTxnsPath, func(c *gin.Context) {
		block, err := DecodeBlockFromHttp(c.Request.Body, w.chain)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = ExecuteTxns(block, w.base, w.land)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		encodeBlockBackHttp(c, block)
	})

	r.POST(FinalizeBlockPath, func(c *gin.Context) {
		block, err := DecodeBlockFromHttp(c.Request.Body, w.chain)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = w.land.RangeList(func(tri tripod.Tripod) error {
			return tri.FinalizeBlock(w.chain, block)
		})
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		encodeBlockBackHttp(c, block)
	})

	r.Run(w.httpPort)
}

func encodeBlockBackHttp(c *gin.Context, b IBlock) {
	byt, err := b.Encode()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	_, err = c.Writer.Write(byt)
	if err != nil {
		logrus.Errorf("write block(%s) back to http error: %s", b.Header().Hash().String(), err.Error())
	}
}
