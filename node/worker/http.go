package worker

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/blockchain"
	. "yu/node"
	"yu/tripod"
)

func (w *Worker) HandleHttp() {
	r := gin.Default()

	r.GET(HeartbeatPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})

	// ------------ Process Block ---------------
	r.POST(StartBlockPath, func(c *gin.Context) {
		block, err := DecodeBlockFromHttp(c.Request.Body, w.chain)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		var needBroadcast bool
		err = w.land.RangeList(func(tri tripod.Tripod) (err error) {
			needBroadcast, err = tri.StartBlock(w.chain, block, w.txPool)
			return
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
