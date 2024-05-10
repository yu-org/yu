package p2p

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/tripod/dev"
	"io"
	"math/rand"
	"os"
	"strconv"
)

const (
	RequestCodeBytesLen = 3
)

// todo: use stream pool to optimize it
type LibP2P struct {
	host      host.Host
	bootNodes []*peerstore.AddrInfo
	pid       protocol.ID
	ps        *pubsub.PubSub
}

func NewP2P(cfg *config.P2pConf) P2pNetwork {
	p2pHost, err := makeP2pHost(cfg)
	if err != nil {
		logrus.Fatal("init p2p-network error: ", err)
	}
	ps, err := pubsub.NewGossipSub(context.Background(), p2pHost)
	if err != nil {
		logrus.Fatal("init p2p gossip error: ", err)
	}
	var bootNodes []*peerstore.AddrInfo
	for _, addrStr := range cfg.Bootnodes {
		addr, err := maddr.NewMultiaddr(addrStr)
		if err != nil {
			logrus.Fatal("new p2p-addr error: ", err)
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			logrus.Fatal("addr info from p2pAddr: ", err)
		}
		bootNodes = append(bootNodes, peer)
	}

	p := &LibP2P{
		host:      p2pHost,
		bootNodes: bootNodes,
		pid:       protocol.ID(cfg.ProtocolID),
		ps:        ps,
	}
	p.AddDefaultTopics()
	return p
}

func (p *LibP2P) LocalID() peerstore.ID {
	return p.host.ID()
}

func (p *LibP2P) LocalIdString() string {
	return p.host.ID().String()
}

func (p *LibP2P) GetBootNodes() (peers []peerstore.ID) {
	for _, addr := range p.bootNodes {
		peers = append(peers, addr.ID)
	}
	return
}

func (p *LibP2P) ConnectBootNodes() error {
	for _, addr := range p.bootNodes {
		err := p.host.Connect(context.Background(), *addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *LibP2P) SetHandlers(handlers map[int]dev.P2pHandler) {
	p.host.SetStreamHandler(p.pid, func(stream network.Stream) {
		go func() {
			var oldErr error
			for {
				err := handleP2pRequest(stream, handlers)
				if err != nil && err != oldErr {
					logrus.Errorf("handle request from node(%s) error: %s",
						stream.Conn().RemotePeer(), err.Error(),
					)
					oldErr = err
				}
			}
		}()
	})
}

func (p *LibP2P) RequestPeer(peerID peerstore.ID, code int, request []byte) ([]byte, error) {
	s, err := p.host.NewStream(context.Background(), peerID, p.pid)
	if err != nil {
		return nil, err
	}
	err = writeToStream(code, request, s)
	if err != nil {
		return nil, err
	}
	return readFromStream(code, s)
}

func (p *LibP2P) PubP2P(topic string, msg []byte) error {
	t, ok := TopicsMap[topic]
	if !ok {
		return yerror.NoP2PTopic
	}
	return t.Publish(context.Background(), msg)
}

func (p *LibP2P) SubP2P(topic string) ([]byte, error) {
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

func handleP2pRequest(s network.Stream, handlers map[int]dev.P2pHandler) error {
	byt, err := readRawStream(s)
	if err != nil {
		return err
	}
	reqCodeByt := byt[:RequestCodeBytesLen]
	reqCode, err := strconv.Atoi(string(reqCodeByt))
	if err != nil {
		return err
	}
	if handler, ok := handlers[reqCode]; ok {
		response, err := handler(byt[RequestCodeBytesLen:])
		if err != nil {
			return err
		}
		return writeToStream(reqCode, response, s)
	}
	return errors.Errorf("no p2p-handler for code(%d)", reqCode)
}

func readFromStream(code int, s network.Stream) ([]byte, error) {
	byt, err := readRawStream(s)
	if err != nil {
		return nil, err
	}
	reqCodeByt := byt[:RequestCodeBytesLen]
	reqCode, err := strconv.Atoi(string(reqCodeByt))
	if err != nil {
		return nil, err
	}
	if reqCode == code {
		return byt[RequestCodeBytesLen:], nil
	}
	return nil, nil
}

func readRawStream(s network.Stream) ([]byte, error) {
	return bufio.NewReader(s).ReadBytes('\n')
}

func writeRawStream(data []byte, s network.Stream) error {
	data = append(data, '\n')
	_, err := s.Write(data)
	return err
}

func writeToStream(code int, data []byte, s network.Stream) error {
	typeBytes := []byte(strconv.Itoa(code))
	data = append(typeBytes, data...)
	return writeRawStream(data, s)
}

func makeP2pHost(cfg *config.P2pConf) (host.Host, error) {
	r, err := loadNodeKeyReader(cfg)
	if err != nil {
		return nil, err
	}
	priv, _, err := crypto.GenerateKeyPairWithReader(cfg.NodeKeyType, cfg.NodeKeyBits, r)
	if err != nil {
		return nil, err
	}
	p2pHost, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(cfg.P2pListenAddrs...),
	)
	if err != nil {
		return nil, err
	}

	hostAddr, err := maddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", p2pHost.ID()))
	if err != nil {
		return nil, err
	}
	addr := p2pHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	logrus.Infof("I am %s", fullAddr)

	return p2pHost, nil
}

func loadNodeKeyReader(cfg *config.P2pConf) (io.Reader, error) {
	if cfg.NodeKey != "" {
		return bytes.NewBufferString(cfg.NodeKey), nil
	}
	if cfg.NodeKeyFile != "" {
		return os.Open(cfg.NodeKeyFile)
	}
	if cfg.NodeKeyRandSeed != 0 {
		return rand.New(rand.NewSource(cfg.NodeKeyRandSeed)), nil
	}
	return nil, nil
}
