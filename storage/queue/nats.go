package queue

import (
	"github.com/nats-io/nats.go"
	"yu/storage"
)

type NatsQueue struct {
	ec *nats.EncodedConn
}

func NewNatsQueue(url, encoder string) (*NatsQueue, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	ec, err := nats.NewEncodedConn(conn, encoder)
	if err != nil {
		return nil, err
	}
	return &NatsQueue{ec: ec}, nil
}

func (nq *NatsQueue) Type() storage.StoreType {
	return storage.Server
}

func (nq *NatsQueue) Kind() storage.StoreKind {
	return storage.Queue
}

func (nq *NatsQueue) Push(topic string, msg interface{}) error {
	err := nq.ec.Publish(topic, msg)
	if err != nil {
		return err
	}
	return nq.ec.Flush()
}

func (nq *NatsQueue) Pop(topic string) (data interface{}, err error) {
	_, err = nq.ec.Subscribe(topic, func(msg interface{}) {
		data = msg
	})
	return
}

func (nq *NatsQueue) PushAsync(topic string, channel interface{}) error {
	return nq.ec.BindSendChan(topic, channel)
}

func (nq *NatsQueue) PopAsync(topic string, channel interface{}) error {
	_, err := nq.ec.BindRecvChan(topic, channel)
	return err
}
