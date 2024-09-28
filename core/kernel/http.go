package kernel

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/protocol"
	"net/http"
)

func (k *Kernel) HandleHttp() {
	r := gin.Default()

	api := r.Group(RootApiPath)
	// POST writing call
	api.POST(WrCallType, func(c *gin.Context) {
		k.handleHttpWr(c)
	})
	// POST reading call
	api.POST(RdCallType, func(c *gin.Context) {
		k.handleHttpRd(c)
	})

	api.GET("receipt", k.GetReceipt)

	api.GET("receipts", k.GetReceipts)
	api.GET("receipts_count", k.GetReceiptsCount)

	if k.cfg.IsAdmin {
		admin := api.Group(AdminType)
		admin.GET("stop", func(c *gin.Context) {
			k.stopChan <- struct{}{}
		})
	}

	err := r.Run(k.httpPort)
	if err != nil {
		logrus.Fatal("serve http failed: ", err)
	}
}

func (k *Kernel) handleHttpWr(c *gin.Context) {
	signedWrCall, err := GetSignedWrCall(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = k.HandleTxn(signedWrCall)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
}

func (k *Kernel) handleHttpRd(c *gin.Context) {
	rdCall, err := GetRdCall(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	respData, err := k.HandleRead(rdCall)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if respData.IsJson {
		c.JSON(respData.StatusCode, respData.DataInterface)
	} else {
		c.Data(respData.StatusCode, respData.ContentType, respData.DataBytes)
	}

	//ctx, err := context.NewReadContext(c, rdCall)
	//if err != nil {
	//	c.AbortWithError(http.StatusBadRequest, err)
	//	return
	//}
	//
	//rd, err := k.Land.GetReading(rdCall.TripodName, rdCall.FuncName)
	//if err != nil {
	//	c.AbortWithError(http.StatusBadRequest, err)
	//	return
	//}
	//rd(ctx)

}
