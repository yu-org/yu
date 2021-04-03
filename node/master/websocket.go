package master

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/common"
	"yu/context"
	. "yu/node"
	. "yu/txn"
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
	var (
		ip   string
		name string
	)
	if m.RunMode == MasterWorker {
		ip, name, err = m.findWorkerIpAndName(tripodName, callName, ExecCall)
		if err != nil {
			BadReqHttpResp(w, FindNoCallStr(tripodName, callName, err))
			return
		}
	}

	err = m.txPool.Insert(name, stxn)
	if err != nil {
		ServerErrorHttpResp(w, err.Error())
		return
	}

	fmap := make(map[string]*TxnsAndWorkerName)
	fmap[ip] = &TxnsAndWorkerName{
		Txns:       []*SignedTxn{stxn},
		WorkerName: name,
	}
	err = m.forwardTxnsForCheck(fmap)
	if err != nil {
		BadReqHttpResp(w, FindNoCallStr(tripodName, callName, err))
		return
	}
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
			BadReqHttpResp(w, FindNoCallStr(qcall.TripodName, qcall.QueryName, err))
			return
		}
		forwardQueryToWorker(ip, w, req)
	} else {
		pubkey, err := GetPubkey(req)
		if err != nil {

			return
		}
		ctx, err := context.NewContext(pubkey.Address(), qcall.Params)
		if err != nil {

			return
		}
		respObj, err := m.land.Query(qcall, ctx)
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
