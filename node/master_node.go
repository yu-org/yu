package node

import (
	"bytes"
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"yu/config"
)

type MasterNode struct {
	ID peer.ID
}

func NewMasterNode(cfg *config.Conf) *MasterNode {
	host, err := makeP2pHost(cfg)
	if err != nil {
		logrus.Panicln("start Master-Node error: ", err)
	}
	return &MasterNode{
		ID: host.ID(),
	}
}

func makeP2pHost(cfg *config.Conf) (host.Host, error) {
	r, err := loadNodeKeyReader(cfg)
	if err != nil {
		return nil, err
	}
	priv, _, err := crypto.GenerateKeyPairWithReader(cfg.NodeKeyType, cfg.NodeKeyBits, r)
	if err != nil {
		return nil, err
	}

	return libp2p.New(
		context.Background(),
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(cfg.P2pAddrs...),
	)
}

func loadNodeKeyReader(cfg *config.Conf) (io.Reader, error) {
	if cfg.NodeKey != "" {
		return bytes.NewBufferString(cfg.NodeKey), nil
	}
	if cfg.NodeKeyFile != "" {
		return os.Open(cfg.NodeKeyFile)
	}
	return rand.Reader, nil
}
