package kernel

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core"
	. "github.com/yu-org/yu/core/chain_env"
	. "github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/subscribe"
	. "github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/tripod/dev"
	. "github.com/yu-org/yu/core/txpool"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/p2p"
	"github.com/yu-org/yu/infra/storage/kv"
	. "github.com/yu-org/yu/utils/ip"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Kernel struct {
	sync.Mutex

	RunMode RunMode
	// Key: NodeKeeper IP, Value: NodeKeeperInfo
	nkDB     kv.KV
	httpPort string
	wsPort   string
	leiLimit uint64

	timeout time.Duration

	chain      types.IBlockChain
	base       types.IBlockBase
	txPool     ItxPool
	stateStore *StateStore

	land *Land

	// event subscription
	sub *subscribe.Subscription

	p2pNetwork p2p.P2pNetwork
}

func NewKernel(
	cfg *KernelConf,
	env *ChainEnv,
	land *Land,
) *Kernel {

	nkDB, err := kv.NewKV(&cfg.NkDB)
	if err != nil {
		logrus.Fatal("init nkDB error: ", err)
	}

	timeout := time.Duration(cfg.Timeout) * time.Second

	m := &Kernel{
		RunMode:    cfg.RunMode,
		leiLimit:   cfg.LeiLimit,
		nkDB:       nkDB,
		timeout:    timeout,
		httpPort:   MakePort(cfg.HttpPort),
		wsPort:     MakePort(cfg.WsPort),
		chain:      env.Chain,
		base:       env.Base,
		txPool:     env.Pool,
		stateStore: env.StateStore,
		sub:        env.Sub,
		p2pNetwork: env.P2pNetwork,

		land: land,
	}

	env.Execute = m.ExecuteTxns

	err = m.InitChain()
	if err != nil {
		logrus.Fatal("init chain error: ", err)
	}

	handerlsMap := make(map[int]dev.P2pHandler, 0)
	handerlsMap[HandshakeCode] = m.handleHsReq
	handerlsMap[SyncTxnsCode] = m.handleSyncTxnsReq

	land.RangeList(func(tri Tripod) error {
		for code, handler := range tri.GetTripodMeta().P2pHandlers {
			handerlsMap[code] = handler
		}
		return nil
	})
	m.p2pNetwork.SetHandlers(handerlsMap)
	err = m.p2pNetwork.ConnectBootNodes()
	if err != nil {
		logrus.Fatal("connect p2p bootnodes error: ", err)
	}

	return m
}

func (m *Kernel) Startup() {

	if len(m.p2pNetwork.GetBootNodes()) > 0 {
		err := m.SyncHistory()
		if err != nil {
			logrus.Fatal("sync history error: ", err)
		}
	}

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

func (m *Kernel) InitChain() error {
	switch m.RunMode {
	case LocalNode:
		return m.land.RangeList(func(tri Tripod) error {
			return tri.InitChain()
		})
	case MasterWorker:
		// todo: init chain

		return nil
	default:
		return NoRunMode
	}
}

//func (m *Kernel) AcceptBlocksFromP2P() error {
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
//			return BlockIllegal(block.Hash)
//		})
//		if err != nil {
//			return err
//		}
//	}
//
//	logrus.Debugf("accept block(%s) height(%d) from p2p", block.Hash.String(), block.Height)
//	return m.chain.InsertBlockFromP2P(block)
//}

func (m *Kernel) AcceptUnpkgTxns() error {
	txns, err := m.subUnpackedTxns()
	if err != nil {
		return err
	}

	switch m.RunMode {
	case MasterWorker:
		//// key: workerIP
		//forwardMap := make(map[string]*TxnsAndWorkerName)
		//for _, txn := range txns {
		//	ecall := txn.GetRaw().Ecall
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
func (m *Kernel) CheckHealth() {
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
//func (m *Kernel) BroadcastTxns() {
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
func (m *Kernel) SyncTxns(block *types.CompactBlock) error {
	txnsHashes := block.TxnsHashes

	needFetch := make([]Hash, 0)
	txns := make(types.SignedTxns, 0)
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
		if m.p2pNetwork.GetBootNodes() == nil {
			fetchPeer = block.PeerID
		} else {
			fetchPeer = m.p2pNetwork.GetBootNodes()[0]
		}

		fetchedTxns, err := m.requestTxns(fetchPeer, block.PeerID, needFetch)
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

		return m.base.SetTxns(block.Hash, fetchedTxns)
	}

	return m.base.SetTxns(block.Hash, txns)
}

func (m *Kernel) SyncHistoryBlocks(blocks []*types.CompactBlock) error {
	switch m.RunMode {
	case LocalNode:
		for _, block := range blocks {
			logrus.Trace("sync history block is ", block.Hash.String())

			err := m.SyncTxns(block)
			if err != nil {
				return err
			}

			err = m.land.RangeList(func(tri Tripod) error {
				if tri.VerifyBlock(block) {
					return nil
				}
				return BlockIllegal(block.Hash)
			})
			if err != nil {
				return err
			}

			err = m.ExecuteTxns(block)
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

func (m *Kernel) registerNodeKeepers(c *gin.Context) {
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

func (m *Kernel) SetNodeKeeper(ip string, info NodeKeeperInfo) error {
	infoByt, err := info.EncodeNodeKeeperInfo()
	if err != nil {
		return err
	}
	return m.nkDB.Set([]byte(ip), infoByt)
}

func (m *Kernel) allNodeKeepersIp() ([]string, error) {
	nkIPs := make([]string, 0)
	err := m.allNodeKeepers(func(ip string, _ *NodeKeeperInfo) error {
		nkIPs = append(nkIPs, ip)
		return nil
	})
	return nkIPs, err
}

func (m *Kernel) WorkersCount() (int, error) {
	count := 0
	err := m.allNodeKeepers(func(_ string, info *NodeKeeperInfo) error {
		count += len(info.WorkersInfo)
		return nil
	})
	return count, err
}

func (m *Kernel) randomGetWorkerIP() (string, error) {
	ips, err := m.allWorkersIP()
	if err != nil {
		return "", err
	}
	randIdx := rand.Intn(len(ips))
	return ips[randIdx], nil
}

func (m *Kernel) allWorkersIP() ([]string, error) {
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
func (m *Kernel) findWorkerIP(tripodName, callName string, callType CallType) (wip string, err error) {
	wip, _, err = m.findWorker(tripodName, callName, callType)
	return
}

func (m *Kernel) findWorkerIpAndName(tripodName, callName string, callType CallType) (wip, name string, err error) {
	var info *WorkerInfo
	wip, info, err = m.findWorker(tripodName, callName, callType)
	if err != nil {
		return
	}
	name = info.Name
	return
}

func (m *Kernel) findWorker(tripodName, callName string, callType CallType) (wip string, wInfo *WorkerInfo, err error) {
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

func (m *Kernel) allNodeKeepers(fn func(nkIP string, info *NodeKeeperInfo) error) error {
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

func (m *Kernel) setNkIfOnline(ip string, isOnline bool) error {
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

func existTxnHash(txnHash Hash, txns []*types.SignedTxn) (*types.SignedTxn, bool) {
	for _, stxn := range txns {
		if stxn.TxnHash == txnHash {
			return stxn, true
		}
	}
	return nil, false
}
