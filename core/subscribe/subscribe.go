package subscribe

import (
	. "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/result"
	"sync"
)

type Subscription struct {
	// key: *websocket.Conn; value: string (topic)
	subscribers sync.Map
	results     map[string]chan *Result
}

func NewSubscription() *Subscription {
	s := &Subscription{
		subscribers: sync.Map{},
		results:     make(map[string]chan *Result),
	}
	go s.emitToClients()
	return s
}

func (s *Subscription) Register(c *Conn, topic string) {
	c.SetCloseHandler(func(_ int, _ string) error {
		s.subscribers.Delete(c)
		return nil
	})

	s.subscribers.Store(c, topic)
}

func (s *Subscription) UnRegister(c *Conn) {
	s.subscribers.Delete(c)
}

func (s *Subscription) Emit(topic string, result *Result) {
	s.results[topic] <- result
}

func (s *Subscription) emitToClients() {
	for topic, resultChan := range s.results {
		select {
		case r := <-resultChan:
			byt, err := r.Encode()
			if err != nil {
				logrus.Errorf("encode Result error: %s", err.Error())
				continue
			}

			s.subscribers.Range(func(connI, topicStr interface{}) bool {
				conn := connI.(*Conn)
				if topic != topicStr.(string) {
					return false
				}
				err = conn.WriteMessage(TextMessage, byt)
				if err != nil {
					logrus.Errorf("emit result to client(%s) error: %s", conn.RemoteAddr().String(), err.Error())
					conn.Close()
					s.subscribers.Delete(connI)
				}
				return true
			})
		}
	}
}
