package master

import (
	"context"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/txn"
	"github.com/yu-org/yu/yerror"
)

type P2PNetwork interface {
	PubToP2P(topic string, msg []byte) error
	SubFromP2P(topic string) ([]byte, error)
}

func (m *Master) initTopics() (err error) {

	topic, err := m.ps.Join(StartBlockTopic)
	if err != nil {
		return
	}
	TopicsMap[StartBlockTopic] = topic
	sub, err := topic.Subscribe()
	if err != nil {
		return
	}
	SubsMap[StartBlockTopic] = sub

	topic, err = m.ps.Join(EndBlockTopic)
	if err != nil {
		return err
	}
	TopicsMap[EndBlockTopic] = topic
	sub, err = topic.Subscribe()
	if err != nil {
		return
	}
	SubsMap[EndBlockTopic] = sub

	topic, err = m.ps.Join(FinalizeBlockTopic)
	if err != nil {
		return
	}
	TopicsMap[FinalizeBlockTopic] = topic
	sub, err = topic.Subscribe()
	if err != nil {
		return
	}
	SubsMap[FinalizeBlockTopic] = sub

	topic, err = m.ps.Join(UnpackedTxnsTopic)
	if err != nil {
		return
	}
	TopicsMap[UnpackedTxnsTopic] = topic
	sub, err = topic.Subscribe()
	if err != nil {
		return
	}
	SubsMap[UnpackedTxnsTopic] = sub

	return
}

func PubToP2P(topic string, msg []byte) error {
	t, ok := TopicsMap[topic]
	if !ok {
		return yerror.NoP2PTopic
	}
	return t.Publish(context.Background(), msg)
}

func SubFromP2P(topic string) ([]byte, error) {
	sub, ok := SubsMap[topic]
	if !ok {
		return nil, yerror.NoP2PTopic
	}
	msg, err := sub.Next(context.Background())
	if err != nil {
		return nil, err
	}
	return msg.Data, nil
}

func (m *Master) pubUnpackedTxns(txns SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return PubToP2P(UnpackedTxnsTopic, byt)
}

func (m *Master) subUnpackedTxns() (SignedTxns, error) {
	byt, err := SubFromP2P(UnpackedTxnsTopic)
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxns(byt)
}

func (m *Master) pubToP2P(topic *pubsub.Topic, msg []byte) error {
	return topic.Publish(context.Background(), msg)
}

func (m *Master) subFromP2P(sub *pubsub.Subscription) ([]byte, error) {
	msg, err := sub.Next(context.Background())
	if err != nil {
		return nil, err
	}
	return msg.Data, nil
}
