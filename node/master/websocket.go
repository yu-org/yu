package master

import (
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/common"
	. "yu/node"
)

func (m *Master) HandleWS() {
	http.HandleFunc(ExecApiPath, func(w http.ResponseWriter, req *http.Request) {
		m.forwardWsCall(w, req, ExecCall)
	})
	http.HandleFunc(QryApiPath, func(w http.ResponseWriter, req *http.Request) {
		m.forwardWsCall(w, req, QryCall)
	})
	logrus.Panic(http.ListenAndServe(m.wsPort, nil))

}
