package p2p

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yu-org/yu/core/tripod/dev"
)

type MockP2p struct {
	nodesNum  int
	topicChan map[string]chan []byte
}

func NewMockP2p(nodesNum int) *MockP2p {
	return &MockP2p{topicChan: make(map[string]chan []byte), nodesNum: nodesNum}
}

func (m *MockP2p) LocalID() peer.ID {
	return peer.ID("")
}

func (m *MockP2p) LocalIdString() string {
	return ""
}

func (m *MockP2p) GetBootNodes() []peer.ID {
	return nil
}

func (m *MockP2p) ConnectBootNodes() error {
	return nil
}

func (m *MockP2p) AddTopic(topicName string) {
	m.topicChan[topicName] = make(chan []byte, m.nodesNum)
}

func (m *MockP2p) SetHandlers(handlers map[int]dev.P2pHandler) {}

func (m *MockP2p) RequestPeer(peerID peer.ID, code int, request []byte) (response []byte, err error) {
	panic("implement me")
}

func (m *MockP2p) PubP2P(topic string, msg []byte) error {
	for i := 0; i < m.nodesNum; i++ {
		m.topicChan[topic] <- msg
	}
	return nil
}

func (m *MockP2p) SubP2P(topic string) ([]byte, error) {
	msg := <-m.topicChan[topic]
	return msg, nil
}
