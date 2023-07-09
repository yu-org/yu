package kernel

import (
	"github.com/gin-gonic/gin"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/types"
	"io"
	"net/http"
)

func (k *Kernel) HandleHttp() {
	r := gin.Default()

	// POST request
	r.POST(WrApiPath, func(c *gin.Context) {
		k.handleHttpWr(c)
	})
	r.POST(RdApiPath, func(c *gin.Context) {
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

	_, err = k.land.GetWriting(stxn.Raw.WrCall)
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
	params, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	rdCall, err := getRdFromHttp(c.Request, string(params))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	switch k.RunMode {
	case LocalNode:
		ctx, err := context.NewReadContext(rdCall.Params)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		err = k.land.Read(rdCall, ctx)
		if err != nil {
			c.String(
				http.StatusInternalServerError,
				err.Error(),
			)
			return
		}
		// FIXME: not only json type.
		c.Data(http.StatusOK, "application/json", ctx.Response())
	}
}
