package master

import (
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/common"
	. "yu/node"
	. "yu/node/handle"
)

func (m *Master) HandleWS() {
	http.HandleFunc(ExecApiPath, func(w http.ResponseWriter, req *http.Request) {
		switch m.RunMode {
		case LocalNode:
			PutWsInTxpool(w, req, m.txPool, m.readyBcTxnsChan)
		case MasterWorker:
			m.forwardWsCall(w, req, ExecCall)
		}
	})

	http.HandleFunc(QryApiPath, func(w http.ResponseWriter, req *http.Request) {
		switch m.RunMode {
		case LocalNode:
			DoWsQryCall(w, req, m.land)
		case MasterWorker:
			m.forwardWsCall(w, req, QryCall)
		}
	})

	logrus.Panic(http.ListenAndServe(m.wsPort, nil))
}
