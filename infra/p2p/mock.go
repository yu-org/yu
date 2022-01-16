package p2p

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/yu-org/yu/core/tripod/dev"
)

type MockP2p struct {
	topicChan map[string]chan []byte
}

func NewMockP2p() *MockP2p {
	return &MockP2p{topicChan: make(map[string]chan []byte)}
}

func (m *MockP2p) LocalID() peer.ID {
	panic("implement me")
}

func (m *MockP2p) LocalIdString() string {
	panic("implement me")
}

func (m *MockP2p) GetBootNodes() []peer.ID {
	panic("implement me")
}

func (m *MockP2p) ConnectBootNodes() error {
	panic("implement me")
}

func (m *MockP2p) AddTopic(topicName string) {
	panic("implement me")
}

func (m *MockP2p) SetHandlers(handlers map[int]dev.P2pHandler) {
	panic("implement me")
}

func (m *MockP2p) RequestPeer(peerID peer.ID, code int, request []byte) (response []byte, err error) {
	panic("implement me")
}

func (m *MockP2p) PubP2P(topic string, msg []byte) error {
	m.topicChan[topic] <- msg
	return nil
}

func (m *MockP2p) SubP2P(topic string) ([]byte, error) {
	msg := <-m.topicChan[topic]
	return msg, nil
}
