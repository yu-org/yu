package kernel

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/protocol"
	"net/http"
)

func (k *Kernel) HandleWS() {
	r := gin.Default()
	r.POST(WrApiPath, func(ctx *gin.Context) {
		k.handleWS(ctx, writing)
	})

	r.GET(RdApiPath, func(ctx *gin.Context) {
		k.handleWS(ctx, reading)
	})

	r.GET(SubResultsPath, func(ctx *gin.Context) {
		k.handleWS(ctx, subscription)
	})
	err := r.Run(k.wsPort)
	if err != nil {
		logrus.Fatal("serve websocket failed: ", err)
	}
}

const (
	reading = iota
	writing
	subscription
)

func (k *Kernel) handleWS(ctx *gin.Context, typ int) {
	upgrade := websocket.Upgrader{}
	c, err := upgrade.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		k.errorAndClose(c, err.Error())
		return
	}
	if typ == subscription {
		logrus.Debugf("Register a Subscription(%s)", c.RemoteAddr().String())
		k.Sub.Register(c)
		return
	}

	_, params, err := c.ReadMessage()
	if err != nil {
		k.errorAndClose(c, fmt.Sprintf("reading websocket message from client error: %v", err))
		return
	}
	switch typ {
	case writing:
		k.handleWsWr(ctx, string(params))
		//case reading:
		//	k.handleWsRd(c, req, string(params))
	}

}

func (k *Kernel) handleWsWr(ctx *gin.Context, params string) {
	signedWrCall, err := GetSignedWrCall(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = k.HandleTxn(signedWrCall)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

//func (k *Kernel) handleWsRd(c *websocket.Conn, req *http.Request, params string) {
//	rdCall, err := getRdFromHttp(req, params)
//	if err != nil {
//		k.errorAndClose(c, fmt.Sprintf("get Reading info from websocket error: %v", err))
//		return
//	}
//
//	switch k.RunMode {
//	case LocalNode:
//		ctx, err := context.NewReadContext(rdCall.Params)
//		if err != nil {
//			k.errorAndClose(c, fmt.Sprintf("new context error: %s", err.Error()))
//			return
//		}
//
//		rd, err := k.Land.GetReading(rdCall)
//		if err != nil {
//			k.errorAndClose(c, err.Error())
//			return
//		}
//		rdErr := rd(ctx)
//		if err != nil {
//			k.errorAndClose(c, rdErr.Error())
//			return
//		}
//		err = c.WriteMessage(websocket.BinaryMessage, ctx.Response())
//		if err != nil {
//			logrus.Errorf("response GetReading receipt error: %s", err.Error())
//		}
//	}
//
//}

func (k *Kernel) errorAndClose(c *websocket.Conn, text string) {
	// FIXME
	c.WriteMessage(websocket.CloseMessage, []byte(text))
	c.Close()
}
