package kernel

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core"
	"net/http"
)

func (k *Kernel) HandleHttp() {
	r := gin.Default()

	// POST request
	r.POST(WrApiPath, func(c *gin.Context) {
		k.handleHttpWr(c)
	})
	// POST request
	r.POST(RdApiPath, func(c *gin.Context) {
		k.handleHttpRd(c)
	})

	if k.cfg.IsAdmin {
		r.GET(StopChainPath, func(c *gin.Context) {
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
