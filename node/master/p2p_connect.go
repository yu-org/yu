package master

import (
	"bytes"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	. "yu/common"
	"yu/config"
	. "yu/node"
	"yu/tripod"
	. "yu/txn"
	. "yu/yerror"
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
	pid := protocol.ID(cfg.ProtocolID)
	m.host.SetStreamHandler(pid, func(s network.Stream) {
		err := m.handleHsReq(s)
		if err != nil {
			logrus.Errorf("handle hand-shake request from node(%s) error: %s",
				s.Conn().RemotePeer().Pretty(), err.Error(),
			)
		}
	})

	for i, addrStr := range cfg.ConnectAddrs {
		addr, err := maddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}

		// sync history block from first connected P2P-node
		if i == 0 {
			s, err := m.host.NewStream(context.Background(), peer.ID, pid)
			if err != nil {
				return err
			}
			err = m.SyncFromP2pNode(s)
			if err != nil {
				return err
			}
		}

		err = m.host.Connect(context.Background(), *peer)
		if err != nil {
			return err
		}
	}
	return nil
}

// Shake hand to the node of p2p network when starts up.
// If we have missing history block, fetch them.
func (m *Master) SyncFromP2pNode(s network.Stream) error {
	resp, err := m.requestP2pNode(nil, s)
	if err != nil {
		return err
	}

	for resp.MissingRange != nil {
		// todo: the missing range maybe very huge and we need fetch them multiple times
		// the remote node will return new Missing blocks-range in this response.
		resp, err = m.requestP2pNode(resp.MissingRange, s)
		if err != nil {
			return err
		}

		if resp.BlocksByt != nil {
			blocks, err := m.chain.DecodeBlocks(resp.BlocksByt)
			if err != nil {
				return err
			}
			err = m.SyncHistoryBlocks(blocks)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (m *Master) requestP2pNode(fetchRange *BlocksRange, s network.Stream) (*HandShakeResp, error) {
	hs, err := m.NewHsReq(fetchRange)
	if err != nil {
		return nil, err
	}
	byt, err := hs.Encode()
	if err != nil {
		return nil, err
	}
	_, err = s.Write(byt)
	if err != nil {
		return nil, err
	}

	respByt, err := ioutil.ReadAll(s)
	if err != nil {
		return nil, err
	}
	return DecodeHsResp(respByt)
}

func (m *Master) handleHsReq(s network.Stream) error {
	byt, err := ioutil.ReadAll(s)
	if err != nil {
		return err
	}

	remoteReq, err := DecodeHsRequest(byt)
	if err != nil {
		return err
	}

	var blocksByt []byte
	if remoteReq.FetchRange != nil {
		blocksByt, err = m.getMissingBlocksByt(remoteReq)
		if err != nil {
			return err
		}
	}

	missingRange, err := m.compareMissingRange(remoteReq.Info)
	if err != nil {
		return err
	}

	hsResp := &HandShakeResp{
		MissingRange: missingRange,
		BlocksByt:    blocksByt,
		// Err: err,
	}
	byt, err = hsResp.Encode()
	if err != nil {
		return err
	}

	_, err = s.Write(byt)
	return err
}

func (m *Master) compareMissingRange(remoteInfo *HandShakeInfo) (*BlocksRange, error) {
	localInfo, err := m.NewHsInfo()
	if err != nil {
		return nil, err
	}
	return localInfo.Compare(m.chain.ConvergeType(), remoteInfo)
}

func (m *Master) getMissingBlocksByt(remoteReq *HandShakeRequest) ([]byte, error) {
	fetchRange := remoteReq.FetchRange
	blocks, err := m.chain.GetRangeBlocks(fetchRange.StartHeight, fetchRange.EndHeight)
	if err != nil {
		return nil, err
	}
	return m.chain.EncodeBlocks(blocks)
}

func (m *Master) AcceptBlocksFromP2P() error {
	block, err := m.subBlock()
	if err != nil {
		return err
	}

	switch m.RunMode {
	case MasterWorker:
		// todo: switch MasterWorker Mode
	case LocalNode:
		err = m.land.RangeList(func(tri tripod.Tripod) error {
			if tri.ValidateBlock(m.chain, block) {
				return nil
			}
			return BlockIllegal(block)
		})
		if err != nil {
			return err
		}
	}

	return m.chain.InsertBlockFromP2P(block)
}

func (m *Master) AcceptUnpkgTxns() error {
	txns, err := m.subUnpackedTxns()
	if err != nil {
		return err
	}

	switch m.RunMode {
	case MasterWorker:
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

		err := m.forwardTxnsForCheck(forwardMap)
		if err != nil {
			return err
		}

		for _, twn := range forwardMap {
			err = m.txPool.BatchInsert(twn.WorkerName, twn.Txns)
			if err != nil {
				return err
			}
		}

	case LocalNode:
		err = m.txPool.BatchInsert("", txns)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Master) forwardTxnsForCheck(forwardMap map[string]*TxnsAndWorkerName) error {
	for workerIP, txns := range forwardMap {
		byt, err := txns.Txns.Encode()
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
