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
			err = w.PutWsInTxpool(c, req)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(fmt.Sprintf("Execution error: %s", err.Error())))
				break
			}
		case QryCall:
			err = w.DoWsQryCall(c, req)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(fmt.Sprintf("Query error: %s", err.Error())))
				break
			}
		}

	}
}

func (w *Worker) PutWsInTxpool(c *websocket.Conn, req *http.Request) error {
	_, msg, err := c.ReadMessage()
	if err != nil {
		return err
	}
	return w.putTxpool(req, JsonString(msg))
}

func (w *Worker) DoWsQryCall(c *websocket.Conn, req *http.Request) error {
	_, msg, err := c.ReadMessage()
	if err != nil {
		return err
	}
	return w.doQryCall(req, JsonString(msg))
}
