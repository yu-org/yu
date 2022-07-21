package kernel

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
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
)

type Kernel struct {
	sync.Mutex

	RunMode RunMode

	httpPort string
	wsPort   string
	leiLimit uint64

	chain   IBlockChain
	base    ItxDB
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

	m := &Kernel{
		RunMode:    cfg.RunMode,
		leiLimit:   cfg.LeiLimit,
		httpPort:   MakePort(cfg.HttpPort),
		wsPort:     MakePort(cfg.WsPort),
		chain:      env.Chain,
		base:       env.TxDB,
		txPool:     env.Pool,
		stateDB:    env.State,
		sub:        env.Sub,
		p2pNetwork: env.P2pNetwork,

		land: land,
	}

	env.Execute = m.ExecuteTxns

	//err := m.InitChain()
	//if err != nil {
	//	logrus.Fatal("init chain error: ", err)
	//}

	handerlsMap := make(map[int]dev.P2pHandler, 0)

	land.RangeList(func(tri *Tripod) error {
		for code, handler := range tri.P2pHandlers {
			handerlsMap[code] = handler
		}
		return nil
	})
	m.p2pNetwork.SetHandlers(handerlsMap)
	err := m.p2pNetwork.ConnectBootNodes()
	if err != nil {
		logrus.Fatal("connect p2p bootnodes error: ", err)
	}

	return m
}

func (m *Kernel) Startup() {

	//if len(m.p2pNetwork.GetBootNodes()) > 0 {
	//	err := m.SyncHistory()
	//	if err != nil {
	//		logrus.Fatal("sync history error: ", err)
	//	}
	//}
	m.land.RangeList(func(tri *Tripod) error {
		tri.InitChain()
		return nil
	})

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

func (m *Kernel) AcceptUnpkgTxns() error {
	txns, err := m.subUnpackedTxns()
	if err != nil {
		return err
	}

	for _, txn := range txns {
		if m.txPool.Exist(txn) {
			continue
		}
		err = m.txPool.CheckTxn(txn)
		if err != nil {
			logrus.Error("check txn from P2P into txpool error: ", err)
			continue
		}
		err = m.txPool.Insert(txn)
		if err != nil {
			logrus.Error("insert txn from P2P into txpool error: ", err)
		}
	}

	return nil
}

func (m *Kernel) subUnpackedTxns() (SignedTxns, error) {
	byt, err := m.p2pNetwork.SubP2P(UnpackedTxnsTopic)
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxns(byt)
}

func (m *Kernel) pubUnpackedTxns(txns SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return m.p2pNetwork.PubP2P(UnpackedTxnsTopic, byt)
}
