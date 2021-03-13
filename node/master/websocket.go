package master

import (
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/common"
	. "yu/node"
	. "yu/node/handle"
	. "yu/utils/error_handle"
)

func (m *Master) HandleWS() {
	http.HandleFunc(ExecApiPath, func(w http.ResponseWriter, req *http.Request) {
		switch m.RunMode {
		case LocalNode:
			PutWsInTxpool(w, req, m.txPool, m.readyBcTxnsChan)
		case MasterWorker:
			m.handleWsCall(w, req, ExecCall)
		}
	})

	http.HandleFunc(QryApiPath, func(w http.ResponseWriter, req *http.Request) {
		switch m.RunMode {
		case LocalNode:
			DoWsQryCall(w, req, m.land)
		case MasterWorker:
			m.handleWsCall(w, req, QryCall)
		}
	})

	logrus.Panic(http.ListenAndServe(m.wsPort, nil))
}

func (m *Master) handleWsCall(w http.ResponseWriter, req *http.Request, callType CallType) {
	tripodName, callName := GetTripodCallName(req)
	ip, err := m.findWorkerIP(tripodName, callName, callType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(BadReqErrStr(tripodName, callName, err)))
		return
	}
	if callType == ExecCall {
		PutWsInTxpool(w, req, m.txPool, m.readyBcTxnsChan)
	}
	w.WriteHeader(http.StatusOK)
}
