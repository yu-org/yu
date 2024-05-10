package p2p

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yu-org/yu/core/tripod/dev"
)

type P2pNetwork interface {
	LocalID() peer.ID
	LocalIdString() string

	GetBootNodes() []peer.ID
	ConnectBootNodes() error

	AddTopic(topicName string)

	SetHandlers(handlers map[int]dev.P2pHandler)
	RequestPeer(peerID peer.ID, code int, request []byte) (response []byte, err error)

	PubP2P(topic string, msg []byte) error
	SubP2P(topic string) ([]byte, error)
}
