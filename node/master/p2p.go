package master

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"yu/config"
	. "yu/node"
)

func makeP2pHost(ctx context.Context, cfg *config.MasterConf) (host.Host, error) {
	r, err := loadNodeKeyReader(cfg)
	if err != nil {
		return nil, err
	}
	priv, _, err := crypto.GenerateKeyPairWithReader(cfg.NodeKeyType, cfg.NodeKeyBits, r)
	if err != nil {
		return nil, err
	}
	return libp2p.New(
		ctx,
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(cfg.P2pListenAddrs...),
	)
}

func loadNodeKeyReader(cfg *config.MasterConf) (io.Reader, error) {
	if cfg.NodeKey != "" {
		return bytes.NewBufferString(cfg.NodeKey), nil
	}
	if cfg.NodeKeyFile != "" {
		return os.Open(cfg.NodeKeyFile)
	}
	return rand.Reader, nil
}

func (m *Master) ConnectP2PNetwork(cfg *config.MasterConf) error {
	m.p2pHost.SetStreamHandler(protocol.ID(cfg.ProtocolID), m.handleStream)

	for _, addrStr := range cfg.ConnectAddrs {
		addr, err := maddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		err = m.p2pHost.Connect(m.ctx, *peer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Master) handleStream(s network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go m.readFromNetwork(rw)
	go m.writeToNetwork(rw)
}

// Read the data of blockchain from P2P network.
func (m *Master) readFromNetwork(rw *bufio.ReadWriter) {
	for {
		byt, err := rw.ReadBytes('\n')
		if err != nil {
			logrus.Errorf("Read data from P2P-network error: %s", err.Error())
			continue
		}
		tbody, err := DecodeTb(byt)
		if err != nil {
			logrus.Errorf("get transfer-body error : %s", err.Error())
			continue
		}

	}
}

// Write and broadcast the data to P2P network.
func (m *Master) writeToNetwork(rw *bufio.ReadWriter) {

	rw.Flush()
}
