package worker

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	. "yu/common"
	. "yu/node"
)

func (w *Worker) HandleWS() {
	http.HandleFunc(ExecApiPath, func(rw http.ResponseWriter, req *http.Request) {
		w.handleWS(rw, req, ExecCall)
	})
	http.HandleFunc(QryApiPath, func(rw http.ResponseWriter, req *http.Request) {
		w.handleWS(rw, req, QryCall)
	})
	logrus.Panic(http.ListenAndServe(w.wsPort, nil))
}

func (w *Worker) handleWS(rw http.ResponseWriter, req *http.Request, callType CallType) {
	tripodName, callName := GetTripodCallName(req)
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(fmt.Sprintf("websocket upgrade error: %s", err.Error())))
		return
	}
	for {
		switch callType {
		case ExecCall:
			err = w.DoWsExecCall(c, tripodName, callName)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(fmt.Sprintf("Execution error: %s", err.Error())))
				break
			}
		case QryCall:
			err = w.DoWsQryCall(c, tripodName, callName)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(fmt.Sprintf("Query error: %s", err.Error())))
				break
			}
		}

	}
}

func (w *Worker) DoWsExecCall(c *websocket.Conn, tripodName, execName string) error {
	var ecallParams EcallParams
	err := c.ReadJSON(&ecallParams)
	if err != nil {
		return err
	}
	ecall := &Ecall{
		TripodName: tripodName,
		ExecName:   execName,
		Params:     ecallParams,
	}
	return w.land.Execute(ecall)
}

func (w *Worker) DoWsQryCall(c *websocket.Conn, tripodName, qryName string) error {
	var qcallParams QcallParams
	err := c.ReadJSON(&qcallParams)
	if err != nil {
		return err
	}
	qcall := &Qcall{
		TripodName: tripodName,
		QueryName:  qryName,
		Params:     qcallParams,
	}
	return w.land.Query(qcall)
}
