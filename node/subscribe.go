package node

import (
	. "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"sync"
	. "yu/result"
)

type Subscription struct {
	// key: *websocket.Conn; value: bool
	subscribers sync.Map
	resultChan  chan Result
}

func NewSubscription() *Subscription {
	s := &Subscription{
		subscribers: sync.Map{},
		resultChan:  make(chan Result, 10),
	}
	go s.emitToClients()
	return s
}

func (s *Subscription) Register(c *Conn) {
	c.SetCloseHandler(func(_ int, _ string) error {
		s.subscribers.Delete(c)
		return nil
	})
	s.subscribers.Store(c, true)
}

func (s *Subscription) Push(result Result) {
	s.resultChan <- result
}

func (s *Subscription) emitToClients() {
	for {
		select {
		case r := <-s.resultChan:
			byt, err := r.Encode()
			if err != nil {
				logrus.Errorf("encode Result error: %s", err.Error())
				continue
			}
			s.subscribers.Range(func(connI, _ interface{}) bool {
				conn := connI.(*Conn)

				err = conn.WriteMessage(BinaryMessage, byt)
				if err != nil {
					logrus.Errorf("emit result to client(%s) error: %s", conn.RemoteAddr().String(), err.Error())
				}
				return true
			})
		}
	}
}
