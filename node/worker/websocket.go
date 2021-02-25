package worker

import (
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/node"
	. "yu/node/handle"
)

func (w *Worker) HandleWS() {
	http.HandleFunc(ExecApiPath, func(rw http.ResponseWriter, req *http.Request) {
		PutWsInTxpool(rw, req, w.txPool)
	})
	http.HandleFunc(QryApiPath, func(rw http.ResponseWriter, req *http.Request) {
		DoWsQryCall(rw, req, w.land)
	})
	logrus.Panic(http.ListenAndServe(w.wsPort, nil))
}
