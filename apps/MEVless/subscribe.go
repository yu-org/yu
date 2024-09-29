package MEVless

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{}

func (m *MEVless) SubscribeOrderCommitment(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("SubscribeOrderCommitment: websocket upgrade failed: %s", err)
		return
	}

	m.wsLock.Lock()
	m.wsClients[c] = true
	m.wsLock.Unlock()

	defer func() {
		_ = c.Close()
		m.wsLock.Lock()
		delete(m.wsClients, c)
		m.wsLock.Unlock()
	}()

	for {
		// Keep client alive
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}

func (m *MEVless) broadcastMessage(message []byte) {
	m.wsLock.Lock()
	defer m.wsLock.Unlock()
	for client := range m.wsClients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			logrus.Error("SubscribeOrderCommitment WriteMessage failed: ", err)
			_ = client.Close()
			delete(m.wsClients, client)
		}
	}
}

func (m *MEVless) StartBroadcasting() {
	for {
		select {
		case oc := <-m.notifyCh:
			byt, err := json.Marshal(oc)
			if err != nil {
				logrus.Error("SubscribeOrderCommitment json.Marshal failed: ", err)
				continue
			}
			m.broadcastMessage(byt)
		}
	}
}

func (m *MEVless) HandleSubscribe() {
	http.HandleFunc("/mev_less", m.SubscribeOrderCommitment)
	go m.StartBroadcasting()
	log.Fatal(http.ListenAndServe(m.cfg.Addr, nil))
}
