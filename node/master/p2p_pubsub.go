package master

import (
	"context"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	. "github.com/yu-altar/yu/txn"
)

const (
	StartBlockTopic    = "start-block"
	EndBlockTopic      = "end-block"
	FinalizeBlockTopic = "finalize-block"
	UnpackedTxnsTopic  = "unpacked-txns"
)

func (m *Master) initTopics() (err error) {
	m.startBlockTopic, err = m.ps.Join(StartBlockTopic)
	if err != nil {
		return
	}
	m.endBlockTopic, err = m.ps.Join(EndBlockTopic)
	if err != nil {
		return err
	}
	m.finalizeBlockTopic, err = m.ps.Join(FinalizeBlockTopic)
	if err != nil {
		return
	}
	m.unpkgTxnsTopic, err = m.ps.Join(UnpackedTxnsTopic)
	if err != nil {
		return
	}

	m.startBlockSub, err = m.startBlockTopic.Subscribe()
	if err != nil {
		return
	}
	m.endBlockSub, err = m.endBlockTopic.Subscribe()
	if err != nil {
		return
	}
	m.finalizeBlockSub, err = m.finalizeBlockTopic.Subscribe()
	if err != nil {
		return
	}
	m.unpackedTxnsSub, err = m.unpkgTxnsTopic.Subscribe()

	return
}

func (m *Master) pubUnpackedTxns(txns SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return m.pubToP2P(m.unpkgTxnsTopic, byt)
}

func (m *Master) subUnpackedTxns() (SignedTxns, error) {
	byt, err := m.subFromP2P(m.unpackedTxnsSub)
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
