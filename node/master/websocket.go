package master

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/common"
	. "yu/node"
	. "yu/utils/error_handle"
)

func (m *Master) HandleWS() {
	http.HandleFunc(ExecApiPath, func(w http.ResponseWriter, req *http.Request) {
		m.handleWS(w, req, ExecCall)
	})

	http.HandleFunc(QryApiPath, func(w http.ResponseWriter, req *http.Request) {
		m.handleWS(w, req, QryCall)
	})

	logrus.Panic(http.ListenAndServe(m.wsPort, nil))
}

func (m *Master) handleWS(w http.ResponseWriter, req *http.Request, callType CallType) {
	upgrade := websocket.Upgrader{}
	c, err := upgrade.Upgrade(w, req, nil)
	if err != nil {
		ServerErrorHttpResp(w, err.Error())
		return
	}
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			BadReqHttpResp(w, err.Error())
			continue
		}
		switch callType {
		case ExecCall:
			m.handleWsExec(w, req, JsonString(msg))
		case QryCall:
			m.handleWsQry(w, req, JsonString(msg))
		}

	}
}

func (m *Master) handleWsExec(w http.ResponseWriter, req *http.Request, params JsonString) {
	tripodName, callName, stxn, err := getExecInfoFromReq(req, params)
	if err != nil {
		BadReqHttpResp(w, err.Error())
		return
	}
	var ip string
	if m.RunMode == MasterWorker {
		ip, err = m.findWorkerIP(tripodName, callName, ExecCall)
		if err != nil {
			BadReqHttpResp(w, BadReqErrStr(tripodName, callName, err))
			return
		}
	}

	// FIXME: insert txn with workerName
	err = m.txPool.Insert(ip, stxn)
	if err != nil {
		ServerErrorHttpResp(w, err.Error())
		return
	}
	// todo: if MasterWorker: forwardTxnsForCheck
	m.readyBcTxnsChan <- stxn
}

func (m *Master) handleWsQry(w http.ResponseWriter, req *http.Request, params JsonString) {
	qcall, err := getQryInfoFromReq(req, params)
	if err != nil {
		BadReqHttpResp(w, err.Error())
		return
	}
	if m.RunMode == MasterWorker {
		ip, err := m.findWorkerIP(qcall.TripodName, qcall.QueryName, QryCall)
		if err != nil {
			BadReqHttpResp(w, BadReqErrStr(qcall.TripodName, qcall.QueryName, err))
			return
		}
		forwardQueryToWorker(ip, w, req)
	} else {
		err = m.land.Query(qcall)
		if err != nil {
			ServerErrorHttpResp(w, BadReqErrStr(qcall.TripodName, qcall.QueryName, err))
			return
		}
	}

}
