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
			HandleWsExec(w, req, m.txPool)
		case MasterWorker:
			m.forwardWsCall(w, req, ExecCall)
		}
	})

	http.HandleFunc(QryApiPath, func(w http.ResponseWriter, req *http.Request) {
		switch m.RunMode {
		case LocalNode:
			HandleWsQry(w, req, m.land)
		case MasterWorker:
			m.forwardWsCall(w, req, QryCall)
		}
	})

	logrus.Panic(http.ListenAndServe(m.wsPort, nil))
}
