package kernel

import (
	"github.com/gin-gonic/gin"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/types"
	"io"
	"net/http"
	"path/filepath"
)

var TryThis = "trythis"

func (k *Kernel) HandleHttp() {
	r := gin.Default()

	// POST request
	r.POST(filepath.Join(WrApiPath, "*path"), func(c *gin.Context) {
		k.handleHttpWr(c)
	})
	// GET request
	r.GET(filepath.Join(RdApiPath, "*path"), func(c *gin.Context) {
		k.handleHttpRd(c)
	})

	r.Run(k.httpPort)
}

func (k *Kernel) handleHttpWr(c *gin.Context) {
	params, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	stxn, err := getWrFromHttp(c.Request, string(params))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	wrCall := stxn.Raw.WrCall
	_, err = k.land.GetWriting(wrCall.TripodName, wrCall.WritingName)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if k.txPool.Exist(stxn) {
		return
	}

	err = k.txPool.CheckTxn(stxn)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	go func() {
		err = k.pubUnpackedTxns(types.FromArray(stxn))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
	}()

	err = k.txPool.Insert(stxn)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

func (k *Kernel) handleHttpRd(c *gin.Context) {
	tripodName, rdName, err := GetTripodCallName(c.Request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	switch k.RunMode {
	case LocalNode:
		ctx, err := context.NewReadContext(c)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		rd, err := k.land.GetReading(tripodName, rdName)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		rd(ctx)
	}
}
