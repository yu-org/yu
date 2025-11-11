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
	p.AddTopic(UnpackedWritingTopic)
}

func (p *LibP2P) AddTopic(topicName string) {
	p.topicMu.RLock()
	if _, exists := p.registeredTopics[topicName]; exists {
		p.topicMu.RUnlock()
		return
	}
	p.topicMu.RUnlock()

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

	p.topicMu.Lock()
	p.registeredTopics[topicName] = struct{}{}
	p.topicMu.Unlock()
}

func (p *LibP2P) HasTopic(topicName string) bool {
	p.topicMu.RLock()
	defer p.topicMu.RUnlock()
	_, ok := p.registeredTopics[topicName]
	return ok
}
