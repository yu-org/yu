package master

import (
	"encoding/json"
	"fmt"
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/context"
	. "github.com/Lawliet-Chan/yu/node"
	. "github.com/Lawliet-Chan/yu/txn"
	. "github.com/Lawliet-Chan/yu/utils/error_handle"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (m *Master) HandleWS() {
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

func (m *Master) handleWS(w http.ResponseWriter, req *http.Request, typ int) {
	upgrade := websocket.Upgrader{}
	c, err := upgrade.Upgrade(w, req, nil)
	if err != nil {
		ServerErrorHttpResp(w, err.Error())
		return
	}
	if typ == subscription {
		logrus.Info("!!!!!!!!  register a sub")
		m.sub.Register(c)
		return
	}

	_, params, err := c.ReadMessage()
	if err != nil {
		BadReqHttpResp(w, fmt.Sprintf("read websocket message from client error: %v", err))
		return
	}
	switch typ {
	case execution:
		m.handleWsExec(w, req, JsonString(params))
	case query:
		m.handleWsQry(c, w, req, JsonString(params))
	}

}

func (m *Master) handleWsExec(w http.ResponseWriter, req *http.Request, params JsonString) {
	_, _, stxn, err := getExecInfoFromReq(req, params)
	if err != nil {
		BadReqHttpResp(w, fmt.Sprintf("get Execution info from websocket error: %v", err))
		return
	}

	switch m.RunMode {
	case MasterWorker:
		//ip, name, err := m.findWorkerIpAndName(tripodName, callName, ExecCall)
		//if err != nil {
		//	BadReqHttpResp(w, FindNoCallStr(tripodName, callName, err))
		//	return
		//}
		//
		//fmap := make(map[string]*TxnsAndWorkerName)
		//fmap[ip] = &TxnsAndWorkerName{
		//	Txns:       FromArray(stxn),
		//	WorkerName: name,
		//}
		//err = m.forwardTxnsForCheck(fmap)
		//if err != nil {
		//	BadReqHttpResp(w, FindNoCallStr(tripodName, callName, err))
		//	return
		//}
		//
		//err = m.txPool.Insert(name, stxn)
		//if err != nil {
		//	ServerErrorHttpResp(w, err.Error())
		//	return
		//}
	case LocalNode:
		err = m.txPool.Insert(stxn)
		if err != nil {
			ServerErrorHttpResp(w, err.Error())
			return
		}
	}

	err = m.pubUnpackedTxns(FromArray(stxn))
	if err != nil {
		BadReqHttpResp(w, fmt.Sprintf("publish Unpacked txn(%s) error: %v", stxn.GetTxnHash().String(), err))
	}
	logrus.Info("publish unpacked txns to P2P")
}

func (m *Master) handleWsQry(c *websocket.Conn, w http.ResponseWriter, req *http.Request, params JsonString) {
	qcall, err := getQryInfoFromReq(req, params)
	if err != nil {
		BadReqHttpResp(w, fmt.Sprintf("get Query info from websocket error: %v", err))
		return
	}

	switch m.RunMode {
	case MasterWorker:
		ip, err := m.findWorkerIP(qcall.TripodName, qcall.QueryName, QryCall)
		if err != nil {
			BadReqHttpResp(w, FindNoCallStr(qcall.TripodName, qcall.QueryName, err))
			return
		}
		forwardQueryToWorker(ip, w, req)
	case LocalNode:
		ctx, err := context.NewContext(NullAddress, qcall.Params)
		if err != nil {
			BadReqHttpResp(w, fmt.Sprintf("new context error: %s", err.Error()))
			return
		}

		respObj, err := m.land.Query(qcall, ctx, m.GetEnv())
		if err != nil {
			ServerErrorHttpResp(w, FindNoCallStr(qcall.TripodName, qcall.QueryName, err))
			return
		}
		respByt, err := json.Marshal(respObj)
		if err != nil {
			ServerErrorHttpResp(w, err.Error())
			return
		}
		err = c.WriteMessage(websocket.BinaryMessage, respByt)
		if err != nil {
			logrus.Errorf("response Query result error: %s", err.Error())
		}
	}

}
