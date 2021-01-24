package master

import (
	"bytes"
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	maddr "github.com/multiformats/go-multiaddr"
	"io"
	"os"
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
)

var MasterInfoKey = []byte("master-info")

type Master struct {
	info    *MasterInfo
	p2pHost host.Host
	metadb  kv.KV
	ctx     context.Context
}

func NewMasterNode(cfg *config.Conf) (*Master, error) {
	metadb, err := kv.NewKV(&cfg.NodeDB)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	p2pHost, err := makeP2pHost(ctx, &cfg.NodeConf)
	if err != nil {
		return nil, err
	}
	data, err := metadb.Get(MasterInfoKey)
	if err != nil {
		return nil, err
	}

	var info *MasterInfo
	if data == nil {
		info = &MasterInfo{
			Name:        cfg.NodeConf.NodeName,
			WorkerNodes: cfg.NodeConf.WorkerNodes,
		}
		infoByt, err := info.EncodeMasterInfo()
		if err != nil {
			return nil, err
		}
		err = metadb.Set(MasterInfoKey, infoByt)
		if err != nil {
			return nil, err
		}
	} else {
		info, err = DecodeMasterInfo(data)
		if err != nil {
			return nil, err
		}
	}

	info.P2pID = p2pHost.ID().String()

	return &Master{
		info,
		p2pHost,
		metadb,
		ctx,
	}, nil
}

func (m *Master) ConnectP2PNetwork(cfg *config.NodeConf) error {
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

func makeP2pHost(ctx context.Context, cfg *config.NodeConf) (host.Host, error) {
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

func loadNodeKeyReader(cfg *config.NodeConf) (io.Reader, error) {
	if cfg.NodeKey != "" {
		return bytes.NewBufferString(cfg.NodeKey), nil
	}
	if cfg.NodeKeyFile != "" {
		return os.Open(cfg.NodeKeyFile)
	}
	return rand.Reader, nil
}
