package kernel

import (
	"encoding/json"
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
	http.HandleFunc(ExecApiPath, func(w http.ResponseWriter, req *http.Request) {
		m.handleWS(w, req, execution)
	})

	http.HandleFunc(QryApiPath, func(w http.ResponseWriter, req *http.Request) {
		m.handleWS(w, req, query)
	})

	http.HandleFunc(SubResultsPath, func(w http.ResponseWriter, req *http.Request) {
		m.handleWS(w, req, subscription)
	})

	logrus.Panic(http.ListenAndServe(m.wsPort, nil))
}

const (
	query = iota
	execution
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
		logrus.Infof("Register a Subscription(%s)", c.RemoteAddr().String())
		m.sub.Register(c)
		return
	}

	_, params, err := c.ReadMessage()
	if err != nil {
		m.errorAndClose(c, fmt.Sprintf("read websocket message from client error: %v", err))
		return
	}
	switch typ {
	case execution:
		m.handleWsExec(c, req, string(params))
	case query:
		m.handleWsQry(c, req, string(params))
	}

}

func (m *Kernel) handleWsExec(c *websocket.Conn, req *http.Request, params string) {
	_, _, stxn, err := getExecInfoFromReq(req, params)
	if err != nil {
		m.errorAndClose(c, fmt.Sprintf("get Execution info from websocket error: %v", err))
		return
	}

	_, _, err = m.land.GetExecLei(stxn.Raw.Ecall)
	if err != nil {
		m.errorAndClose(c, err.Error())
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

func (m *Kernel) handleWsQry(c *websocket.Conn, req *http.Request, params string) {
	qcall, err := getQryInfoFromReq(req, params)
	if err != nil {
		m.errorAndClose(c, fmt.Sprintf("get Query info from websocket error: %v", err))
		return
	}

	switch m.RunMode {
	case LocalNode:
		ctx, err := context.NewContext(NullAddress, qcall.Params)
		if err != nil {
			m.errorAndClose(c, fmt.Sprintf("new context error: %s", err.Error()))
			return
		}

		respObj, err := m.land.Query(qcall, ctx)
		if err != nil {
			m.errorAndClose(c, FindNoCallStr(qcall.TripodName, qcall.QueryName, err))
			return
		}
		respByt, err := json.Marshal(respObj)
		if err != nil {
			m.errorAndClose(c, err.Error())
			return
		}
		err = c.WriteMessage(websocket.BinaryMessage, respByt)
		if err != nil {
			logrus.Errorf("response Query result error: %s", err.Error())
		}
	}

}

func (m *Kernel) errorAndClose(c *websocket.Conn, text string) {
	logrus.Error(text)
	c.WriteMessage(websocket.CloseMessage, []byte(text))
}
