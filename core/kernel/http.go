package kernel

import (
	"github.com/gin-gonic/gin"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/types"
	"io/ioutil"
	"net/http"
)

func (m *Kernel) HandleHttp() {
	r := gin.Default()

	// POST request
	r.POST(WrApiPath, func(c *gin.Context) {
		m.handleHttpWr(c)
	})
	r.POST(RdApiPath, func(c *gin.Context) {
		m.handleHttpRd(c)
	})

	r.Run(m.httpPort)
}

func (m *Kernel) handleHttpWr(c *gin.Context) {
	params, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	_, _, stxn, err := getWrInfoFromReq(c.Request, string(params))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	_, err = m.land.GetWriting(stxn.Raw.WrCall)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if m.txPool.Exist(stxn) {
		return
	}

	err = m.txPool.CheckTxn(stxn)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	go func() {
		err = m.pubUnpackedTxns(types.FromArray(stxn))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
	}()

	err = m.txPool.Insert(stxn)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

func (m *Kernel) handleHttpRd(c *gin.Context) {
	params, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	qcall, err := getRdInfoFromReq(c.Request, string(params))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	switch m.RunMode {
	case LocalNode:
		ctx, err := context.NewReadContext(qcall.Params)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		err = m.land.Read(qcall, ctx)
		if err != nil {
			c.String(
				http.StatusBadRequest,
				FindNoCallStr(qcall.TripodName, qcall.ReadingName, err),
			)
			return
		}
		// FIXME: not only json type.
		c.JSON(http.StatusOK, ctx.Response())
	}
}
