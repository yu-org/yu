package kernel

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/chain_env"
	. "github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/subscribe"
	. "github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/tripod/dev"
	. "github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/p2p"
	. "github.com/yu-org/yu/utils/ip"
	"sync"
	"time"
)

type Kernel struct {
	sync.Mutex

	RunMode RunMode

	httpPort string
	wsPort   string
	leiLimit uint64

	timeout time.Duration

	chain   IBlockChain
	base    IyuDB
	txPool  ItxPool
	stateDB IState

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

	timeout := time.Duration(cfg.Timeout) * time.Second

	m := &Kernel{
		RunMode:    cfg.RunMode,
		leiLimit:   cfg.LeiLimit,
		timeout:    timeout,
		httpPort:   MakePort(cfg.HttpPort),
		wsPort:     MakePort(cfg.WsPort),
		chain:      env.Chain,
		base:       env.YuDB,
		txPool:     env.Pool,
		stateDB:    env.State,
		sub:        env.Sub,
		p2pNetwork: env.P2pNetwork,

		land: land,
	}

	env.Execute = m.ExecuteTxns

	err := m.InitChain()
	if err != nil {
		logrus.Fatal("init chain error: ", err)
	}

	handerlsMap := make(map[int]dev.P2pHandler, 0)
	handerlsMap[HandshakeCode] = m.handleHsReq
	handerlsMap[SyncTxnsCode] = m.handleSyncTxnsReq

	land.RangeList(func(tri Tripod) error {
		for code, handler := range tri.GetTripodHeader().P2pHandlers {
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

	go m.HandleHttp()
	go m.HandleWS()

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
		errs := m.txPool.BatchInsert(txns)
		for _, e := range errs {
			logrus.Error("insert txn of P2P into txpool error: ", e)
		}
	}

	return nil
}

// SyncTxns sync txns of P2P-network
func (m *Kernel) SyncTxns(block *CompactBlock) error {
	txnsHashes := block.TxnsHashes

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

func (m *Kernel) SyncHistoryBlocks(blocks []*CompactBlock) error {
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

func existTxnHash(txnHash Hash, txns []*SignedTxn) (*SignedTxn, bool) {
	for _, stxn := range txns {
		if stxn.TxnHash == txnHash {
			return stxn, true
		}
	}
	return nil, false
}
