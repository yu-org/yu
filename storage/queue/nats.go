package queue

import (
	"github.com/nats-io/nats.go"
	"github.com/yu-org/yu/storage"
)

type NatsQueue struct {
	nc *nats.Conn
	ec *nats.EncodedConn
}

func NewNatsQueue(url, encoder string) (*NatsQueue, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	ec, err := nats.NewEncodedConn(nc, encoder)
	if err != nil {
		return nil, err
	}
	return &NatsQueue{nc: nc, ec: ec}, nil
}

func (nq *NatsQueue) Type() storage.StoreType {
	return storage.Server
}

func (nq *NatsQueue) Kind() storage.StoreKind {
	return storage.Queue
}

func (nq *NatsQueue) Push(topic string, msg []byte) error {
	err := nq.nc.Publish(topic, msg)
	if err != nil {
		return err
	}
	return nq.ec.Flush()
}

func (nq *NatsQueue) Pop(topic string) (data []byte, err error) {
	_, err = nq.nc.Subscribe(topic, func(msg *nats.Msg) {
		data = msg.Data
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
