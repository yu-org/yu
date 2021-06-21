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
	for {
		_, params, err := c.ReadMessage()
		if err != nil {
			BadReqHttpResp(w, err.Error())
			continue
		}
		switch typ {
		case execution:
			m.handleWsExec(w, req, JsonString(params))
		case query:
			m.handleWsQry(w, req, JsonString(params))
		}

	}
}

func (m *Master) handleWsExec(w http.ResponseWriter, req *http.Request, params JsonString) {
	tripodName, callName, stxn, err := getExecInfoFromReq(req, params)
	if err != nil {
		BadReqHttpResp(w, err.Error())
		return
	}

	switch m.RunMode {
	case MasterWorker:
		ip, name, err := m.findWorkerIpAndName(tripodName, callName, ExecCall)
		if err != nil {
			BadReqHttpResp(w, FindNoCallStr(tripodName, callName, err))
			return
		}

		fmap := make(map[string]*TxnsAndWorkerName)
		fmap[ip] = &TxnsAndWorkerName{
			Txns:       FromArray(stxn),
			WorkerName: name,
		}
		err = m.forwardTxnsForCheck(fmap)
		if err != nil {
			BadReqHttpResp(w, FindNoCallStr(tripodName, callName, err))
			return
		}

		err = m.txPool.Insert(name, stxn)
		if err != nil {
			ServerErrorHttpResp(w, err.Error())
			return
		}
	case LocalNode:
		err = m.txPool.Insert("", stxn)
		if err != nil {
			ServerErrorHttpResp(w, err.Error())
			return
		}
	}

	err = m.pubUnpackedTxns(FromArray(stxn))
	if err != nil {
		BadReqHttpResp(w, err.Error())
	}
	logrus.Info("publish unpacked txns to P2P")
}

func (m *Master) handleWsQry(w http.ResponseWriter, req *http.Request, params JsonString) {
	qcall, err := getQryInfoFromReq(req, params)
	if err != nil {
		BadReqHttpResp(w, err.Error())
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
		pubkey, err := GetPubkey(req)
		if err != nil {
			BadReqHttpResp(w, fmt.Sprintf("get pubkey error: %s", err.Error()))
			return
		}
		ctx, err := context.NewContext(pubkey.Address(), qcall.Params)
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
		_, err = w.Write(respByt)
		if err != nil {
			logrus.Errorf("response Query result error: %s", err.Error())
		}
	}

}
