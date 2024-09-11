package kernel

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/env"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/tripod/dev"
	. "github.com/yu-org/yu/core/types"
	. "github.com/yu-org/yu/utils/ip"
	"sync"
)

type Kernel struct {
	mutex sync.Mutex

	cfg *KernelConf

	RunMode RunMode

	stopChan chan struct{}

	httpPort string
	wsPort   string
	leiLimit uint64

	*ChainEnv

	Land *Land
}

func NewKernel(
	cfg *KernelConf,
	env *ChainEnv,
	land *Land,
) *Kernel {
	k := &Kernel{
		cfg:      cfg,
		RunMode:  cfg.RunMode,
		stopChan: make(chan struct{}),
		leiLimit: cfg.LeiLimit,
		httpPort: MakePort(cfg.HttpPort),
		wsPort:   MakePort(cfg.WsPort),
		ChainEnv: env,
		Land:     land,
	}

	env.Execute = k.OrderedExecute

	// Configure the handlers in P2P network

	handlersMap := make(map[int]P2pHandler, 0)

	land.RangeList(func(tri *Tripod) error {
		for code, handler := range tri.P2pHandlers {
			handlersMap[code] = handler
		}
		return nil
	})
	k.P2pNetwork.SetHandlers(handlersMap)

	// connect the P2P network
	err := k.P2pNetwork.ConnectBootNodes()
	if err != nil {
		logrus.Fatal("connect p2p bootnodes error: ", err)
	}

	return k
}

func (k *Kernel) WithExecuteFn(fn ExecuteFn) {
	k.Execute = fn
}

func (k *Kernel) Startup() {
	k.InitBlockChain()

	go k.HandleHttp()
	go k.HandleWS()

	k.Run()
}

func (k *Kernel) Stop() {
	k.stopChan <- struct{}{}
}

func (k *Kernel) InitBlockChain() {
	genesisBlock := k.makeGenesisBlock()
	k.Land.RangeList(func(tri *Tripod) error {
		tri.Init.InitChain(genesisBlock)
		return nil
	})
}

func (k *Kernel) AcceptUnpkgTxns() error {
	txns, err := k.subUnpackedTxns()
	if err != nil {
		return err
	}

	for _, txn := range txns {
		if k.CheckReplayAttack(txn) {
			continue
		}
		txn.FromP2P = true

		logrus.WithField("p2p", "accept-txn").
			Tracef("txn(%s) from network, content: %v", txn.TxnHash.String(), txn.Raw.WrCall)

		err = k.Pool.CheckTxn(txn)
		if err != nil {
			logrus.Error("check txn from P2P into txpool error: ", err)
			continue
		}
		err = k.Pool.Insert(txn)
		if err != nil {
			logrus.Error("insert txn from P2P into txpool error: ", err)
		}
	}

	return nil
}

func (k *Kernel) subUnpackedTxns() (SignedTxns, error) {
	byt, err := k.P2pNetwork.SubP2P(UnpackedTxnsTopic)
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
	return k.P2pNetwork.PubP2P(UnpackedTxnsTopic, byt)
}

func (k *Kernel) GetTripodInstance(name string) any {
	return k.Land.GetTripodInstance(name)
}

func (k *Kernel) GetTxn(txnHash Hash) (txn *SignedTxn, err error) {
	txn, _ = k.Pool.GetTxn(txnHash)
	if txn == nil {
		txn, err = k.TxDB.GetTxn(txnHash)
	}
	return
}
