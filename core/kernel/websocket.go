package kernel

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/types"
	"net/http"
)

func (m *Kernel) HandleWS() {
	http.HandleFunc(WrApiPath, func(w http.ResponseWriter, req *http.Request) {
		m.handleWS(w, req, writing)
	})

	http.HandleFunc(RdApiPath, func(w http.ResponseWriter, req *http.Request) {
		m.handleWS(w, req, reading)
	})

	http.HandleFunc(SubResultsPath, func(w http.ResponseWriter, req *http.Request) {
		m.handleWS(w, req, subscription)
	})

	logrus.Panic(http.ListenAndServe(m.wsPort, nil))
}

const (
	reading = iota
	writing
	subscription
)

func (m *Kernel) handleWS(w http.ResponseWriter, req *http.Request, typ int) {
	upgrade := websocket.Upgrader{}
	c, err := upgrade.Upgrade(w, req, nil)
	if err != nil {
		m.errorAndClose(c, err.Error())
		return
	}
	if typ == subscription {
		logrus.Debugf("Register a Subscription(%s)", c.RemoteAddr().String())
		m.sub.Register(c)
		return
	}

	_, params, err := c.ReadMessage()
	if err != nil {
		m.errorAndClose(c, fmt.Sprintf("reading websocket message from client error: %v", err))
		return
	}
	switch typ {
	case writing:
		m.handleWsWr(c, req, string(params))
	case reading:
		m.handleWsRd(c, req, string(params))
	}

}

func (m *Kernel) handleWsWr(c *websocket.Conn, req *http.Request, params string) {
	stxn, err := getWrInfoFromReq(req, params)
	if err != nil {
		m.errorAndClose(c, fmt.Sprintf("get Writing info from websocket error: %v", err))
		return
	}

	_, err = m.land.GetWriting(stxn.Raw.WrCall)
	if err != nil {
		m.errorAndClose(c, err.Error())
		return
	}

	if m.txPool.Exist(stxn) {
		return
	}

	err = m.txPool.CheckTxn(stxn)
	if err != nil {
		m.errorAndClose(c, err.Error())
		return
	}

	go func() {
		err = m.pubUnpackedTxns(types.FromArray(stxn))
		if err != nil {
			m.errorAndClose(c, fmt.Sprintf("publish Unpacked txn(%s) error: %v", stxn.TxnHash.String(), err))
		}
	}()

	err = m.txPool.Insert(stxn)
	if err != nil {
		m.errorAndClose(c, err.Error())
		return
	}
}

func (m *Kernel) handleWsRd(c *websocket.Conn, req *http.Request, params string) {
	qcall, err := getRdInfoFromReq(req, params)
	if err != nil {
		m.errorAndClose(c, fmt.Sprintf("get Reading info from websocket error: %v", err))
		return
	}

	switch m.RunMode {
	case LocalNode:
		ctx, err := context.NewReadContext(qcall.Params)
		if err != nil {
			m.errorAndClose(c, fmt.Sprintf("new context error: %s", err.Error()))
			return
		}

		err = m.land.Read(qcall, ctx)
		if err != nil {
			m.errorAndClose(c, FindNoCallStr(qcall.TripodName, qcall.ReadingName, err))
			return
		}
		err = c.WriteMessage(websocket.BinaryMessage, ctx.Response())
		if err != nil {
			logrus.Errorf("response Read result error: %s", err.Error())
		}
	}

}

func (m *Kernel) errorAndClose(c *websocket.Conn, text string) {
	// FIXEME
	c.WriteMessage(websocket.CloseMessage, []byte(text))
	c.Close()
}
