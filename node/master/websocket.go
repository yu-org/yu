package master

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/common"
	. "yu/node"
)

func (m *Master) HandleWS() {
	r := mux.NewRouter()
	r.PathPrefix()
	http.HandleFunc(ExecApiWsPath, func(w http.ResponseWriter, req *http.Request) {
		m.forwardWsCall(w, req, ExecCall)
	})
	http.HandleFunc(QryApiWsPath, func(w http.ResponseWriter, req *http.Request) {
		m.forwardWsCall(w, req, QryCall)
	})
	logrus.Panic(http.ListenAndServe(m.wsPort, nil))

}
