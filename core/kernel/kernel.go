package kernel

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/env"
	. "github.com/yu-org/yu/core/state"
	. "github.com/yu-org/yu/core/subscribe"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/tripod/dev"
	. "github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
	. "github.com/yu-org/yu/infra/p2p"
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
	txDB    ItxDB
	txPool  ItxPool
	stateDB IState

	land *Land

	// event subscription
	sub *Subscription

	p2pNetwork P2pNetwork
}

func NewKernel(
	cfg *KernelConf,
	env *ChainEnv,
	land *Land,
	executeFn ExecuteFn,
) *Kernel {

	k := &Kernel{
		RunMode:    cfg.RunMode,
		leiLimit:   cfg.LeiLimit,
		httpPort:   MakePort(cfg.HttpPort),
		wsPort:     MakePort(cfg.WsPort),
		chain:      env.Chain,
		txDB:       env.TxDB,
		txPool:     env.Pool,
		stateDB:    env.State,
		sub:        env.Sub,
		p2pNetwork: env.P2pNetwork,

		land: land,
	}

	if executeFn == nil {
		env.Execute = k.OrderedExecute
	} else {
		env.Execute = executeFn
	}

	// Configure the handlers in P2P network

	handlersMap := make(map[int]P2pHandler, 0)

	land.RangeList(func(tri *Tripod) error {
		for code, handler := range tri.P2pHandlers {
			handlersMap[code] = handler
		}
		return nil
	})
	k.p2pNetwork.SetHandlers(handlersMap)

	// connect the P2P network
	err := k.p2pNetwork.ConnectBootNodes()
	if err != nil {
		logrus.Fatal("connect p2p bootnodes error: ", err)
	}

	return k
}

func (k *Kernel) Startup() {
	k.InitChain()

	// TODO: need to abstract out as handleTxn(*SignedTxn)
	go k.HandleHttp()
	go k.HandleWS()

	k.Run()
}

func (k *Kernel) InitChain() {
	k.land.RangeList(func(tri *Tripod) error {
		tri.InitChain()
		return nil
	})
}

func (k *Kernel) AcceptUnpkgTxns() error {
	txns, err := k.subUnpackedTxns()
	if err != nil {
		return err
	}

	for _, txn := range txns {
		if k.txPool.Exist(txn) {
			continue
		}
		txn.FromP2P = true

		logrus.WithField("p2p", "accept-txn").
			Tracef("txn(%s) from network, content: %v", txn.TxnHash.String(), txn.Raw.WrCall)

		err = k.txPool.CheckTxn(txn)
		if err != nil {
			logrus.Error("check txn from P2P into txpool error: ", err)
			continue
		}
		err = k.txPool.Insert(txn)
		if err != nil {
			logrus.Error("insert txn from P2P into txpool error: ", err)
		}
	}

	return nil
}

func (k *Kernel) subUnpackedTxns() (SignedTxns, error) {
	byt, err := k.p2pNetwork.SubP2P(UnpackedTxnsTopic)
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxns(byt)
}

func (k *Kernel) pubUnpackedTxns(txns SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return k.p2pNetwork.PubP2P(UnpackedTxnsTopic, byt)
}
