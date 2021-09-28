package master

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/node"
	. "github.com/yu-org/yu/state"
	"github.com/yu-org/yu/storage/kv"
	"github.com/yu-org/yu/subscribe"
	. "github.com/yu-org/yu/tripod"
	. "github.com/yu-org/yu/txn"
	. "github.com/yu-org/yu/txpool"
	. "github.com/yu-org/yu/utils/ip"
	. "github.com/yu-org/yu/yerror"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Master struct {
	sync.Mutex

	host           host.Host
	ps             *pubsub.PubSub
	protocolID     protocol.ID
	ConnectedPeers []peer.ID

	RunMode RunMode
	// Key: NodeKeeper IP, Value: NodeKeeperInfo
	nkDB     kv.KV
	httpPort string
	wsPort   string
	leiLimit uint64

	timeout time.Duration

	chain      IBlockChain
	base       IBlockBase
	txPool     ItxPool
	stateStore *StateStore

	land *Land

	// event subscription
	sub *subscribe.Subscription

	PubP2P func(topic string, msg []byte) error
	SubP2P func(topic string) ([]byte, error)

	// P2P topic
	//startBlockTopic    *pubsub.Topic
	//endBlockTopic      *pubsub.Topic
	//finalizeBlockTopic *pubsub.Topic
	//unpkgTxnsTopic     *pubsub.Topic
	//
	//// P2P topic subscribe
	//startBlockSub    *pubsub.Subscription
	//endBlockSub      *pubsub.Subscription
	//finalizeBlockSub *pubsub.Subscription
	//unpackedTxnsSub  *pubsub.Subscription
	//
	//msgOnStart    chan []byte
	//msgOnEnd      chan []byte
	//msgOnFinalize chan []byte
}

func NewMaster(
	cfg *MasterConf,
	chain IBlockChain,
	base IBlockBase,
	txPool ItxPool,
	land *Land,
) (*Master, error) {
	var err error
	if chain == nil {
		chain, err = NewBlockChain(&cfg.BlockChain)
		if err != nil {
			logrus.Panicf("load blockchain error: %s", err.Error())
		}
	}
	if base == nil {
		base, err = NewBlockBase(&cfg.BlockBase)
		if err != nil {
			logrus.Panicf("load blockbase error: %s", err.Error())
		}
	}
	stateStore, err := NewStateStore(&cfg.State)
	if err != nil {
		logrus.Panicf("load stateKV error: %s", err.Error())
	}

	if txPool == nil {
		txPool = LocalWithDefaultChecks(&cfg.Txpool)
	}

	nkDB, err := kv.NewKV(&cfg.NkDB)
	if err != nil {
		return nil, err
	}
	pid := protocol.ID(cfg.ProtocolID)
	ctx := context.Background()
	p2pHost, err := makeP2pHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, p2pHost)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(cfg.Timeout) * time.Second

	m := &Master{
		host:       p2pHost,
		ps:         ps,
		protocolID: pid,
		RunMode:    cfg.RunMode,
		leiLimit:   cfg.LeiLimit,
		nkDB:       nkDB,
		timeout:    timeout,
		httpPort:   MakePort(cfg.HttpPort),
		wsPort:     MakePort(cfg.WsPort),
		chain:      chain,
		base:       base,
		txPool:     txPool,
		stateStore: stateStore,

		land: land,
		sub:  subscribe.NewSubscription(),

		PubP2P: PubToP2P,
		SubP2P: SubFromP2P,
	}
	err = m.initTopics()
	if err != nil {
		return nil, err
	}
	err = m.InitChain()
	if err != nil {
		return nil, err
	}

	err = m.ConnectP2PNetwork(cfg)
	return m, err
}

func (m *Master) P2pID() string {
	return m.host.ID().String()
}

func (m *Master) Startup() {

	if m.RunMode == MasterWorker {
		go m.CheckHealth()
	}

	go m.HandleHttp()
	go m.HandleWS()

	//go func() {
	//	for {
	//		err := m.AcceptBlocksFromP2P()
	//		if err != nil {
	//			logrus.Errorf("accept blocks error: %s", err.Error())
	//		}
	//	}
	//
	//}()

	go func() {
		for {
			err := m.AcceptUnpkgTxns()
			if err != nil {
				logrus.Errorf("accept unpacked txns error: %s", err.Error())
			}
		}

	}()

	m.Run()
}

func (m *Master) InitChain() error {
	switch m.RunMode {
	case LocalNode:
		return m.land.RangeList(func(tri Tripod) error {
			return tri.InitChain(m.GetEnv(), m.land)
		})
	case MasterWorker:
		// todo: init chain

		return nil
	default:
		return NoRunMode
	}
}

//func (m *Master) AcceptBlocksFromP2P() error {
//	block, err := m.subBlock()
//	if err != nil {
//		return err
//	}
//
//	switch m.RunMode {
//	case MasterWorker:
//		// todo: switch MasterWorker Mode
//	case LocalNode:
//		err = m.land.RangeList(func(tri Tripod) error {
//			if tri.VerifyBlock(block, m.GetEnv()) {
//				return nil
//			}
//			return BlockIllegal(block.GetHash())
//		})
//		if err != nil {
//			return err
//		}
//	}
//
//	logrus.Debugf("accept block(%s) height(%d) from p2p", block.GetHash().String(), block.GetHeight())
//	return m.chain.InsertBlockFromP2P(block)
//}

func (m *Master) AcceptUnpkgTxns() error {
	txns, err := m.subUnpackedTxns()
	if err != nil {
		return err
	}

	switch m.RunMode {
	case MasterWorker:
		//// key: workerIP
		//forwardMap := make(map[string]*TxnsAndWorkerName)
		//for _, txn := range txns {
		//	ecall := txn.GetRaw().GetEcall()
		//	tripodName := ecall.TripodName
		//	execName := ecall.ExecName
		//	workerIP, workerName, err := m.findWorkerIpAndName(tripodName, execName, ExecCall)
		//	if err != nil {
		//		return err
		//	}
		//	oldTxns := forwardMap[workerIP].Txns
		//	forwardMap[workerIP] = &TxnsAndWorkerName{
		//		Txns:       append(oldTxns, txn),
		//		WorkerName: workerName,
		//	}
		//}
		//
		//err := m.forwardTxnsForCheck(forwardMap)
		//if err != nil {
		//	return err
		//}
		//
		//for _, twn := range forwardMap {
		//	err = m.txPool.BatchInsert(twn.WorkerName, twn.Txns)
		//	if err != nil {
		//		return err
		//	}
		//}

	case LocalNode:
		err = m.txPool.BatchInsert(txns)
		if err != nil {
			return err
		}
	}

	return nil
}

// Check the health of NodeKeepers by SendHeartbeat to them.
func (m *Master) CheckHealth() {
	for {
		nkAddrs, err := m.allNodeKeepersIp()
		if err != nil {
			logrus.Errorf("get all NodeKeepers error: %s", err.Error())
		}
		SendHeartbeats(
			nkAddrs,
			func(ip string) error {
				return m.setNkIfOnline(ip, true)
			},
			func(ip string) error {
				return m.setNkIfOnline(ip, false)
			})
		time.Sleep(m.timeout)
	}
}

// FIXME: when number of txns is just less than NumOfBcTxns
//func (m *Master) BroadcastTxns() {
//	var txns SignedTxns
//	for {
//		select {
//		case txn := <-m.readyBcTxnsChan:
//			txns = append(txns, txn)
//			if len(txns) == m.NumOfBcTxns {
//				body, err := NewTxnsTransferBody(txns)
//				if err != nil {
//					logrus.Errorf("new TxnTransferBody error: %s", err.Error())
//					continue
//				}
//				m.txnsBcChan <- body
//				txns = nil
//			}
//		}
//	}
//}

// sync txns of P2P-network
func (m *Master) SyncTxns(block IBlock) error {
	txnsHashes := block.GetTxnsHashes()

	needFetch := make([]Hash, 0)
	txns := make(SignedTxns, 0)
	for _, txnHash := range txnsHashes {
		stxn, err := m.txPool.GetTxn(txnHash)
		if err != nil {
			return err
		}
		if stxn == nil {
			logrus.Infof("need fetch packed-txn(%s)", txnHash.String())
			needFetch = append(needFetch, txnHash)
		} else {
			txns = append(txns, stxn)
		}
	}

	if len(needFetch) > 0 {
		logrus.Info(" start sub packed txns")

		var fetchPeer peer.ID
		if m.ConnectedPeers == nil {
			fetchPeer = block.GetPeerID()
		} else {
			fetchPeer = m.ConnectedPeers[0]
		}

		fetchedTxns, err := m.requestTxns(fetchPeer, block.GetPeerID(), needFetch)
		if err != nil {
			return err
		}

		for _, txnHash := range needFetch {
			_, exist := existTxnHash(txnHash, fetchedTxns)
			if !exist {
				return NoTxnInP2P(txnHash)
			}
		}

		for _, fetchedTxn := range fetchedTxns {
			err = m.txPool.NecessaryCheck(fetchedTxn)
			if err != nil {
				return err
			}
		}

		return m.base.SetTxns(block.GetHash(), fetchedTxns)
	}

	return m.base.SetTxns(block.GetHash(), txns)
}

func (m *Master) SyncHistoryBlocks(blocks []IBlock) error {
	switch m.RunMode {
	case LocalNode:
		for _, block := range blocks {
			logrus.Trace("sync history block is ", block.GetHash().String())

			err := m.SyncTxns(block)
			if err != nil {
				return err
			}

			err = m.land.RangeList(func(tri Tripod) error {
				if tri.VerifyBlock(block, m.GetEnv()) {
					return nil
				}
				return BlockIllegal(block.GetHash())
			})
			if err != nil {
				return err
			}

			err = ExecuteTxns(block, m.GetEnv(), m.land)
			if err != nil {
				return err
			}

			err = m.chain.AppendBlock(block)
			if err != nil {
				return err
			}
		}
		return nil
	case MasterWorker:
		// todo
		return nil
	default:
		return NoRunMode
	}
}

func (m *Master) GetEnv() *ChainEnv {
	return &ChainEnv{
		StateStore: m.stateStore,
		RunMode:    m.RunMode,
		Chain:      m.chain,
		Base:       m.base,
		Pool:       m.txPool,
		P2pID:      m.host.ID(),
		Sub:        m.sub,
		PubP2P:     PubToP2P,
		SubP2P:     SubFromP2P,
	}
}

func (m *Master) registerNodeKeepers(c *gin.Context) {
	m.Lock()
	defer m.Unlock()
	var nkInfo NodeKeeperInfo
	err := c.ShouldBindJSON(&nkInfo)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("NodeKeeperInfo decode failed: %s", err.Error()),
		)
		return
	}
	nkIP := MakeIp(c.ClientIP(), nkInfo.ServesPort)

	err = m.SetNodeKeeper(nkIP, nkInfo)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("store NodeKeeper(%s) error: %s", nkIP, err.Error()),
		)
		return
	}

	c.String(http.StatusOK, "")
	logrus.Infof("NodeKeeper(%s) register succeed!", nkIP)
}

func (m *Master) SetNodeKeeper(ip string, info NodeKeeperInfo) error {
	infoByt, err := info.EncodeNodeKeeperInfo()
	if err != nil {
		return err
	}
	return m.nkDB.Set([]byte(ip), infoByt)
}

func (m *Master) allNodeKeepersIp() ([]string, error) {
	nkIPs := make([]string, 0)
	err := m.allNodeKeepers(func(ip string, _ *NodeKeeperInfo) error {
		nkIPs = append(nkIPs, ip)
		return nil
	})
	return nkIPs, err
}

func (m *Master) WorkersCount() (int, error) {
	count := 0
	err := m.allNodeKeepers(func(_ string, info *NodeKeeperInfo) error {
		count += len(info.WorkersInfo)
		return nil
	})
	return count, err
}

func (m *Master) randomGetWorkerIP() (string, error) {
	ips, err := m.allWorkersIP()
	if err != nil {
		return "", err
	}
	randIdx := rand.Intn(len(ips))
	return ips[randIdx], nil
}

func (m *Master) allWorkersIP() ([]string, error) {
	var workersIP []string
	err := m.allNodeKeepers(func(_ string, info *NodeKeeperInfo) error {
		for ip, _ := range info.WorkersInfo {
			workersIP = append(workersIP, ip)
		}
		return nil
	})
	return workersIP, err
}

// find workerIP by Execution/Query name
func (m *Master) findWorkerIP(tripodName, callName string, callType CallType) (wip string, err error) {
	wip, _, err = m.findWorker(tripodName, callName, callType)
	return
}

func (m *Master) findWorkerIpAndName(tripodName, callName string, callType CallType) (wip, name string, err error) {
	var info *WorkerInfo
	wip, info, err = m.findWorker(tripodName, callName, callType)
	if err != nil {
		return
	}
	name = info.Name
	return
}

func (m *Master) findWorker(tripodName, callName string, callType CallType) (wip string, wInfo *WorkerInfo, err error) {
	err = m.allNodeKeepers(func(nkIP string, info *NodeKeeperInfo) error {
		if !info.Online {
			return NodeKeeperDead(nkIP)
		}
		for workerIp, workerInfo := range info.WorkersInfo {
			if !workerInfo.Online {
				return WorkerDead(workerInfo.Name)
			}
			triInfo, ok := workerInfo.TripodsInfo[tripodName]
			if !ok {
				return TripodNotFound(tripodName)
			}
			var callArr []string
			switch callType {
			case ExecCall:
				callArr = triInfo.ExecNames
			case QryCall:
				callArr = triInfo.QueryNames
			}
			for _, call := range callArr {
				if call == callName {
					wip = workerIp
					wInfo = &workerInfo
					return nil
				}
			}
		}
		return nil
	})
	if err != nil {
		return
	}
	if wip == "" || wInfo == nil {
		switch callType {
		case ExecCall:
			err = ExecNotFound(callName)
		case QryCall:
			err = QryNotFound(callName)
		}
	}
	return
}

func (m *Master) allNodeKeepers(fn func(nkIP string, info *NodeKeeperInfo) error) error {
	iter, err := m.nkDB.Iter(nil)
	if err != nil {
		return err
	}
	defer iter.Close()
	for iter.Valid() {
		ipByt, infoByt, err := iter.Entry()
		if err != nil {
			return err
		}
		ip := string(ipByt)
		info, err := DecodeNodeKeeperInfo(infoByt)
		if err != nil {
			return err
		}
		err = fn(ip, info)
		if err != nil {
			return err
		}
		err = iter.Next()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Master) setNkIfOnline(ip string, isOnline bool) error {
	tx, err := m.nkDB.NewKvTxn()
	if err != nil {
		return err
	}
	info, err := getNkWithTx(tx, ip)
	if err != nil {
		return err
	}
	if info.Online != isOnline {
		info.Online = isOnline
		err = setNkWithTx(tx, ip, info)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func getNkWithTx(txn kv.KvTxn, ip string) (*NodeKeeperInfo, error) {
	infoByt, err := txn.Get([]byte(ip))
	if err != nil {
		return nil, err
	}
	return DecodeNodeKeeperInfo(infoByt)
}

func setNkWithTx(txn kv.KvTxn, ip string, info *NodeKeeperInfo) error {
	infoByt, err := info.EncodeNodeKeeperInfo()
	if err != nil {
		return err
	}
	return txn.Set([]byte(ip), infoByt)
}

func existTxnHash(txnHash Hash, txns []*SignedTxn) (*SignedTxn, bool) {
	for _, stxn := range txns {
		if stxn.GetTxnHash() == txnHash {
			return stxn, true
		}
	}
	return nil, false
}
