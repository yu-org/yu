package handle

//import (
//	"fmt"
//	"github.com/gorilla/websocket"
//	"net/http"
//	. "yu/common"
//	"yu/tripod"
//	"yu/txn"
//	"yu/txpool"
//)
//
//func PutWsInTxpool(rw http.ResponseWriter, req *http.Request, txPool txpool.ItxPool, broadcastChan chan<- txn.IsignedTxn) {
//	upgrader := websocket.Upgrader{}
//	c, err := upgrader.Upgrade(rw, req, nil)
//	if err != nil {
//		rw.WriteHeader(http.StatusInternalServerError)
//		rw.Write([]byte(fmt.Sprintf("websocket upgrade error: %s", err.Error())))
//		return
//	}
//	for {
//		_, msg, err := c.ReadMessage()
//		if err != nil {
//			rw.WriteHeader(http.StatusInternalServerError)
//			rw.Write([]byte(fmt.Sprintf("Read websocket msg error: %s", err.Error())))
//			break
//		}
//		err = putTxpool(req, JsonString(msg), txPool, broadcastChan)
//		if err != nil {
//			rw.WriteHeader(http.StatusInternalServerError)
//			rw.Write([]byte(fmt.Sprintf("Execution error: %s", err.Error())))
//			break
//		}
//	}
//}
//
//func DoWsQryCall(rw http.ResponseWriter, req *http.Request, land *tripod.Land) {
//	upgrader := websocket.Upgrader{}
//	c, err := upgrader.Upgrade(rw, req, nil)
//	if err != nil {
//		rw.WriteHeader(http.StatusInternalServerError)
//		rw.Write([]byte(fmt.Sprintf("websocket upgrade error: %s", err.Error())))
//		return
//	}
//	for {
//		_, msg, err := c.ReadMessage()
//		if err != nil {
//			rw.WriteHeader(http.StatusInternalServerError)
//			rw.Write([]byte(fmt.Sprintf("Read websocket msg error: %s", err.Error())))
//			break
//		}
//		err = doQryCall(req, JsonString(msg), land)
//		if err != nil {
//			rw.WriteHeader(http.StatusInternalServerError)
//			rw.Write([]byte(fmt.Sprintf("Query error: %s", err.Error())))
//			break
//		}
//	}
//}
