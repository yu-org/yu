package p2p

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	. "github.com/yu-org/yu/common"
)

var (
	TopicsMap = make(map[string]*pubsub.Topic, 0)
	SubsMap   = make(map[string]*pubsub.Subscription, 0)
)

func (p *LibP2P) AddDefaultTopics() {
	p.AddTopic(StartBlockTopic)
	p.AddTopic(EndBlockTopic)
	p.AddTopic(FinalizeBlockTopic)
	p.AddTopic(UnpackedTxnsTopic)
}

func (p *LibP2P) AddTopic(topicName string) {
	topic, err := p.ps.Join(topicName)
	if err != nil {
		return
	}
	TopicsMap[topicName] = topic
	sub, err := topic.Subscribe()
	if err != nil {
		return
	}
	SubsMap[topicName] = sub
}
