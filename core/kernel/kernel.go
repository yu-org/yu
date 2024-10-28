package kernel

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/env"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/tripod/dev"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/utils/ip"
	"sync"
)

type Kernel struct {
	mutex sync.Mutex

	cfg *config.KernelConf

	RunMode common.RunMode

	stopChan chan struct{}

	httpPort string
	wsPort   string
	leiLimit uint64

	*env.ChainEnv

	Land *tripod.Land
}

func NewKernel(
	cfg *config.KernelConf,
	env *env.ChainEnv,
	land *tripod.Land,
) *Kernel {
	k := &Kernel{
		cfg:      cfg,
		RunMode:  cfg.RunMode,
		stopChan: make(chan struct{}),
		leiLimit: cfg.LeiLimit,
		httpPort: ip.MakePort(cfg.HttpPort),
		wsPort:   ip.MakePort(cfg.WsPort),
		ChainEnv: env,
		Land:     land,
	}

	env.Execute = k.OrderedExecute

	// Configure the handlers in P2P network

	handlersMap := make(map[int]dev.P2pHandler, 0)

	land.RangeList(func(tri *tripod.Tripod) error {
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

func (k *Kernel) WithExecuteFn(fn env.ExecuteFn) {
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
	k.Land.RangeList(func(tri *tripod.Tripod) error {
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

func (k *Kernel) subUnpackedTxns() (types.SignedTxns, error) {
	byt, err := k.P2pNetwork.SubP2P(common.UnpackedTxnsTopic)
	if err != nil {
		return nil, err
	}
	return types.DecodeSignedTxns(byt)
}

func (k *Kernel) pubUnpackedTxns(txns types.SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return k.P2pNetwork.PubP2P(common.UnpackedTxnsTopic, byt)
}

func (k *Kernel) GetTripodInstance(name string) any {
	return k.Land.GetTripodInstance(name)
}

func (k *Kernel) GetTxn(txnHash common.Hash) (txn *types.SignedTxn, err error) {
	txn, _ = k.Pool.GetTxn(txnHash)
	if txn == nil {
		txn, err = k.TxDB.GetTxn(txnHash)
	}
	return
}

// WithBronzes fill the bronzes, if you have bronzes, must use it, and then WithTripods.
func (k *Kernel) WithBronzes(bronzeInstances ...any) *Kernel {
	bronzes := make([]*tripod.Bronze, 0)
	for _, v := range bronzeInstances {
		bronzes = append(bronzes, tripod.ResolveBronze(v))
	}

	for i, t := range bronzes {
		t.SetChainEnv(k.ChainEnv)
		t.SetLand(k.Land)
		t.SetInstance(bronzeInstances[i])
	}

	k.Land.SetBronzes(bronzes...)

	for _, bronzeInstance := range bronzeInstances {
		err := tripod.InjectToBronze(k.Land, bronzeInstance)
		if err != nil {
			logrus.Fatal("inject bronze failed: ", err)
		}
	}
	return k
}

func (k *Kernel) WithTripods(tripodInstances ...any) *Kernel {
	tripods := make([]*tripod.Tripod, 0)
	for _, v := range tripodInstances {
		tripods = append(tripods, tripod.ResolveTripod(v))
	}

	for i, t := range tripods {
		t.SetChainEnv(k.ChainEnv)
		t.SetLand(k.Land)
		t.SetInstance(tripodInstances[i])
	}

	k.Land.SetTripods(tripods...)

	for _, tri := range tripods {
		k.Pool.WithTripodCheck(tri.Name(), tri.TxnChecker)
	}

	for _, tripodInstance := range tripodInstances {
		err := tripod.InjectToTripod(tripodInstance)
		if err != nil {
			logrus.Fatal("inject tripod failed: ", err)
		}
	}
	return k
}
