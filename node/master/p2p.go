package master

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	. "yu/common"
	"yu/config"
	. "yu/node"
	"yu/tripod"
	. "yu/txn"
	. "yu/yerror"
)

type P2pInfo struct {
	host  host.Host
	ps    *pubsub.PubSub
	ctx   context.Context
	topic *pubsub.Topic
}

func makeP2pHost(ctx context.Context, cfg *config.MasterConf) (host.Host, error) {
	r, err := loadNodeKeyReader(cfg)
	if err != nil {
		return nil, err
	}
	priv, _, err := crypto.GenerateKeyPairWithReader(cfg.NodeKeyType, cfg.NodeKeyBits, r)
	if err != nil {
		return nil, err
	}
	p2pHost, err := libp2p.New(
		ctx,
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(cfg.P2pListenAddrs...),
	)
	if err != nil {
		return nil, err
	}

	hostAddr, err := maddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", p2pHost.ID().Pretty()))
	if err != nil {
		return nil, err
	}
	addr := p2pHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	logrus.Infof("I am %s", fullAddr)

	return p2pHost, nil
}

func loadNodeKeyReader(cfg *config.MasterConf) (io.Reader, error) {
	if cfg.NodeKey != "" {
		return bytes.NewBufferString(cfg.NodeKey), nil
	}
	if cfg.NodeKeyFile != "" {
		return os.Open(cfg.NodeKeyFile)
	}
	return nil, nil
}

func (m *Master) ConnectP2PNetwork(cfg *config.MasterConf) error {
	m.p2pInfo.host.SetStreamHandler(protocol.ID(cfg.ProtocolID), m.handleStream)

	for _, addrStr := range cfg.ConnectAddrs {
		addr, err := maddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		err = m.p2pInfo.host.Connect(m.p2pInfo.ctx, *peer)
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
			logrus.Errorf("decode data from P2P-network error: %s", err.Error())
			continue
		}

		err = m.handleTransferBody(tbody)
		if err != nil {
			logrus.Errorf("handle transfer-body error：%s", err.Error())
		}
	}
}

// Write and broadcast the data to P2P network.
func (m *Master) writeToNetwork(rw *bufio.ReadWriter) {
	for {
		select {
		case blocksBody := <-m.blockBcChan:
			byt, err := blocksBody.Encode()
			if err != nil {
				logrus.Errorf("encode block-body error: %s", err.Error())
				continue
			}
			_, err = rw.Write(byt)
			if err != nil {
				logrus.Errorf("write block-body to P2P network error: %s", err.Error())
				continue
			}
			rw.Flush()
		case txnsBody := <-m.txnsBcChan:
			byt, err := txnsBody.Encode()
			if err != nil {
				logrus.Errorf("encode txns-body error: %s", err.Error())
				continue
			}
			_, err = rw.Write(byt)
			if err != nil {
				logrus.Errorf("write txns-body error: %s", err.Error())
				continue
			}
			rw.Flush()
		}
	}
}

func (m *Master) handleTransferBody(tbody *TransferBody) error {
	switch tbody.Type {
	case BlockTransfer:
		block, err := tbody.DecodeBlockBody()
		if err != nil {
			return err
		}
		err = m.land.RangeList(func(tri tripod.Tripod) error {
			if tri.ValidateBlock(block) {
				return nil
			}
			return BlockIllegal(block.GetHeader().GetHash())
		})
		if err != nil {
			return err
		}
		return m.chain.InsertBlockFromP2P(block)
	case TxnsTransfer:
		txns, err := tbody.DecodeTxnsBody()
		if err != nil {
			return err
		}

		// key: workerIP
		forwardMap := make(map[string]*TxnsAndWorkerName)
		for _, txn := range txns {
			ecall := txn.GetRaw().Ecall()
			tripodName := ecall.TripodName
			execName := ecall.ExecName
			workerIP, workerName, err := m.findWorkerIpAndName(tripodName, execName, ExecCall)
			if err != nil {
				return err
			}
			oldTxns := forwardMap[workerIP].Txns
			forwardMap[workerIP] = &TxnsAndWorkerName{
				Txns:       append(oldTxns, txn),
				WorkerName: workerName,
			}
		}

		if m.RunMode == MasterWorker {
			err := m.forwardTxnsForCheck(forwardMap)
			if err != nil {
				return err
			}
		}

		for _, twn := range forwardMap {
			err = m.txPool.BatchInsert(twn.WorkerName, twn.Txns)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return NoTransferBodyType
	}
}

func (m *Master) forwardTxnsForCheck(forwardMap map[string]*TxnsAndWorkerName) error {
	for workerIP, txns := range forwardMap {
		newTbody, err := NewTxnsTransferBody(txns.Txns)
		if err != nil {
			return err
		}
		byt, err := newTbody.Encode()
		if err != nil {
			return err
		}
		_, err = PostRequest(workerIP+CheckTxnsPath, byt)
		if err != nil {
			return err
		}
	}

	return nil
}

type TxnsAndWorkerName struct {
	Txns       SignedTxns
	WorkerName string
}

func (m *Master) pubToP2P(blockHash Hash, txns SignedTxns) error {
	topic, err := m.p2pInfo.ps.Join(blockHash.String())
	if err != nil {
		return err
	}
	m.p2pInfo.topic = topic
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return m.p2pInfo.topic.Publish(m.p2pInfo.ctx, byt)
}

func (m *Master) subFromP2P(blockHash Hash) ([]*SignedTxn, error) {
	topic, err := m.p2pInfo.ps.Join(blockHash.String())
	if err != nil {
		return nil, err
	}
	m.p2pInfo.topic = topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}
	msg, err := sub.Next(m.p2pInfo.ctx)
	if err != nil {
		return nil, err
	}
	stxns, err := m.txPool.NewEmptySignedTxns().Decode(msg.Data)
	if err != nil {
		return nil, err
	}
	return stxns.ToArray(), nil
}

func (m *Master) closeTopic() error {
	if m.p2pInfo.topic != nil {
		return m.p2pInfo.topic.Close()
	}
	return nil
}
