package master

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/config"
	. "github.com/Lawliet-Chan/yu/node"
	. "github.com/Lawliet-Chan/yu/txn"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"os"
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
	if cfg.NodeKeyRandSeed != 0 {
		return rand.New(rand.NewSource(cfg.NodeKeyRandSeed)), nil
	}
	return nil, nil
}

func (m *Master) ConnectP2PNetwork(cfg *config.MasterConf) error {
	pid := protocol.ID(cfg.ProtocolID)
	m.host.SetStreamHandler(pid, func(s network.Stream) {

		go func() {
			var oldErr error
			for {
				err := m.handleHsReq(s)
				if err != nil && err != oldErr {
					logrus.Errorf("handle hand-shake request from node(%s) error: %s",
						s.Conn().RemotePeer().Pretty(), err.Error(),
					)
					oldErr = err
				}
			}
		}()

	})

	ctx := context.Background()

	for i, addrStr := range cfg.ConnectAddrs {
		addr, err := maddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}

		err = m.host.Connect(ctx, *peer)
		if err != nil {
			return err
		}

		// todo: we need make some strategy to choose the best node(s)
		// sync history block from first connected P2P-node
		if i == 0 {
			s, err := m.host.NewStream(ctx, peer.ID, pid)
			if err != nil {
				return err
			}
			err = m.SyncFromP2pNode(s)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Shake hand to the node of p2p network when starts up.
// If we have missing history block, fetch them.
func (m *Master) SyncFromP2pNode(s network.Stream) error {
	logrus.Info("start to sync history from other node")

	resp, err := m.requestP2pNode(nil, s)
	if err != nil {
		return err
	}
	if resp.Err != nil {
		return resp.Err
	}

	for resp.MissingRange != nil {
		// todo: the missing range maybe very huge and we need fetch them multiple times
		// the remote node will return new Missing blocks-range in this response.
		resp, err = m.requestP2pNode(resp.MissingRange, s)
		if err != nil {
			return err
		}

		if resp.Err != nil {
			return resp.Err
		}

		if resp.BlocksByt != nil {
			blocks, err := m.chain.DecodeBlocks(resp.BlocksByt)
			if err != nil {
				return err
			}

			logrus.Info("fetch history blocks are: ", blocks)

			err = m.SyncHistoryBlocks(blocks)
			if err != nil {
				return err
			}

			resp.MissingRange = nil
		}

		if resp.TxnsByt != nil {
			for blockHash, byt := range resp.TxnsByt {
				txns, err := DecodeSignedTxns(byt)
				if err != nil {
					return err
				}
				err = m.base.SetTxns(blockHash, txns)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func (m *Master) handleHsReq(s network.Stream) error {

	byt, err := readFromStream(s)
	if err != nil {
		return err
	}

	remoteReq, err := DecodeHsRequest(byt)
	if err != nil {
		return err
	}

	logrus.Info("remote info genesis-block is: ", remoteReq.Info.GenesisBlockHash.String())
	logrus.Info("remote info end-block hash is ", remoteReq.Info.EndBlockHash.String())
	logrus.Info("remote request fetch range: ", remoteReq.FetchRange)

	var (
		blocksByt []byte
		txnsByt   map[Hash][]byte
	)
	if remoteReq.FetchRange != nil {
		blocksByt, txnsByt, err = m.getMissingBlocksTxns(remoteReq)
		if err != nil {
			return err
		}
	}

	missingRange, err := m.compareMissingRange(remoteReq.Info)

	if missingRange != nil {
		logrus.Infof("missing range start-height is %d,  end-height is %d", missingRange.StartHeight, missingRange.EndHeight)
	}

	hsResp := &HandShakeResp{
		MissingRange: missingRange,
		BlocksByt:    blocksByt,
		TxnsByt:      txnsByt,
		Err:          err,
	}
	byt, err = hsResp.Encode()
	if err != nil {
		return err
	}

	return writeToStream(byt, s)
}

func (m *Master) requestP2pNode(fetchRange *BlocksRange, s network.Stream) (*HandShakeResp, error) {
	hs, err := m.NewHsReq(fetchRange)
	if err != nil {
		return nil, err
	}

	logrus.Info("handshake info genesis-blockHash is   ", hs.Info.GenesisBlockHash.String())
	logrus.Info("handshake info end-blockhash is    ", hs.Info.EndBlockHash.String())
	logrus.Info("handshake fetch range request is     ", hs.FetchRange)

	byt, err := hs.Encode()
	if err != nil {
		return nil, err
	}

	err = writeToStream(byt, s)
	if err != nil {
		return nil, err
	}

	respByt, err := readFromStream(s)
	if err != nil {
		return nil, err
	}
	return DecodeHsResp(respByt)
}

func (m *Master) compareMissingRange(remoteInfo *HandShakeInfo) (*BlocksRange, error) {
	localInfo, err := m.NewHsInfo()
	if err != nil {
		return nil, err
	}
	return localInfo.Compare(remoteInfo)
}

func (m *Master) getMissingBlocksTxns(remoteReq *HandShakeRequest) ([]byte, map[Hash][]byte, error) {
	fetchRange := remoteReq.FetchRange
	blocks, err := m.chain.GetRangeBlocks(fetchRange.StartHeight, fetchRange.EndHeight)
	if err != nil {
		return nil, nil, err
	}
	blocksByt, err := m.chain.EncodeBlocks(blocks)
	if err != nil {
		return nil, nil, err
	}

	txnsByt := make(map[Hash][]byte)
	for _, block := range blocks {
		blockHash := block.GetHash()
		txns, err := m.base.GetTxns(blockHash)
		if err != nil {
			return nil, nil, err
		}
		byt, err := FromArray(txns...).Encode()
		if err != nil {
			return nil, nil, err
		}
		txnsByt[blockHash] = byt
	}

	return blocksByt, txnsByt, nil
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

func readFromStream(s network.Stream) ([]byte, error) {
	return bufio.NewReader(s).ReadBytes('\n')
}

func writeToStream(data []byte, s network.Stream) error {
	data = append(data, '\n')
	_, err := s.Write(data)
	return err
}
