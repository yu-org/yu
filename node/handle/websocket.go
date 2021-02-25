package handle

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	. "yu/common"
	"yu/tripod"
	"yu/txpool"
)

func HandleWsExec(rw http.ResponseWriter, req *http.Request, txPool txpool.ItxPool) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(fmt.Sprintf("websocket upgrade error: %s", err.Error())))
		return
	}
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("Read websocket msg error: %s", err.Error())))
			break
		}
		err = putTxpool(req, JsonString(msg), txPool)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("Execution error: %s", err.Error())))
			break
		}
	}
}

func HandleWsQry(rw http.ResponseWriter, req *http.Request, land *tripod.Land) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(fmt.Sprintf("websocket upgrade error: %s", err.Error())))
		return
	}
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("Read websocket msg error: %s", err.Error())))
			break
		}
		err = doQryCall(req, JsonString(msg), land)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("Query error: %s", err.Error())))
			break
		}
	}
}
