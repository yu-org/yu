package p2p

import (
	. "github.com/yu-org/yu/common"
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
	sub, err := topic.Subscribe()
	if err != nil {
		return
	}

	p.topicMu.Lock()
	p.registeredTopics[topicName] = topic
	p.subscriptions[topicName] = sub
	p.topicMu.Unlock()
}

func (p *LibP2P) HasTopic(topicName string) bool {
	p.topicMu.RLock()
	defer p.topicMu.RUnlock()
	_, ok := p.registeredTopics[topicName]
	return ok
}
