package kernel

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core"
	"github.com/yu-org/yu/core/types"
	"net/http"
	"path/filepath"
)

func (k *Kernel) HandleWS() {
	r := gin.Default()
	r.POST(filepath.Join(WrApiPath, "*path"), func(ctx *gin.Context) {
		k.handleWS(ctx.Writer, ctx.Request, writing)
	})

	r.GET(filepath.Join(RdApiPath, "*path"), func(ctx *gin.Context) {
		k.handleWS(ctx.Writer, ctx.Request, reading)
	})

	r.GET(SubResultsPath, func(ctx *gin.Context) {
		k.handleWS(ctx.Writer, ctx.Request, subscription)
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

func (k *Kernel) handleWS(w http.ResponseWriter, req *http.Request, typ int) {
	upgrade := websocket.Upgrader{}
	c, err := upgrade.Upgrade(w, req, nil)
	if err != nil {
		k.errorAndClose(c, err.Error())
		return
	}
	if typ == subscription {
		logrus.Debugf("Register a Subscription(%s)", c.RemoteAddr().String())
		k.sub.Register(c)
		return
	}

	_, params, err := c.ReadMessage()
	if err != nil {
		k.errorAndClose(c, fmt.Sprintf("reading websocket message from client error: %v", err))
		return
	}
	switch typ {
	case writing:
		k.handleWsWr(c, req, string(params))
		//case reading:
		//	k.handleWsRd(c, req, string(params))
	}

}

func (k *Kernel) handleWsWr(c *websocket.Conn, req *http.Request, params string) {
	stxn, err := getWrFromHttp(req, params)
	if err != nil {
		k.errorAndClose(c, fmt.Sprintf("get Writing info from websocket error: %v", err))
		return
	}

	wrCall := stxn.Raw.WrCall
	_, err = k.land.GetWriting(wrCall.TripodName, wrCall.WritingName)
	if err != nil {
		k.errorAndClose(c, err.Error())
		return
	}

	if k.txPool.Exist(stxn) {
		return
	}

	err = k.txPool.CheckTxn(stxn)
	if err != nil {
		k.errorAndClose(c, err.Error())
		return
	}

	go func() {
		err = k.pubUnpackedTxns(types.FromArray(stxn))
		if err != nil {
			k.errorAndClose(c, fmt.Sprintf("publish Unpacked txn(%s) error: %v", stxn.TxnHash.String(), err))
		}
	}()

	err = k.txPool.Insert(stxn)
	if err != nil {
		k.errorAndClose(c, err.Error())
		return
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
//		rd, err := k.land.GetReading(rdCall)
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
//			logrus.Errorf("response GetReading result error: %s", err.Error())
//		}
//	}
//
//}

func (k *Kernel) errorAndClose(c *websocket.Conn, text string) {
	// FIXEME
	c.WriteMessage(websocket.CloseMessage, []byte(text))
	c.Close()
}
