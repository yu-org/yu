package queue

import (
	"yu/config"
	"yu/storage"
	. "yu/yerror"
)

type Queue interface {
	storage.StorageType
	Push(topic string, msg []byte) error
	Pop(topic string) ([]byte, error)

	// The type of msgChan must be chan!
	PushAsync(topic string, msgChan interface{}) error
	PopAsync(topic string, msgChan interface{}) error
}

func NewQueue(cfg *config.QueueConf) (Queue, error) {
	switch cfg.QueueType {
	case "nats":
		return NewNatsQueue(cfg.Url, cfg.Encoder)

	default:
		return nil, NoQueueType
	}
}
