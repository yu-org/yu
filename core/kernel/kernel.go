package kernel

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/env"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/tripod/dev"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/utils/ip"
)

type Kernel struct {
	cfg *config.KernelConf

	RunMode common.RunMode

	stopChan chan struct{}

	httpPort string
	wsPort   string
	leiLimit uint64

	*env.ChainEnv

	Land *tripod.Land
	wg   *sync.WaitGroup
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
		wg:       &sync.WaitGroup{},
	}

	env.Execute = k.SeqExecuteWritings

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

func (k *Kernel) WaitExit() {
	k.wg.Wait()
}

func (k *Kernel) Startup() {
	k.InitBlockChain()

	go k.HandleHttp()
	go k.HandleWS()

	k.wg.Add(1)
	go k.AcceptUnpkgTxnsJob()
	go k.Run()
}

func (k *Kernel) Stop() {
	close(k.stopChan)
	k.wg.Wait()
}

func (k *Kernel) InitBlockChain() {
	genesisBlock := k.makeGenesisBlock()
	k.Land.RangeList(func(tri *tripod.Tripod) error {
		tri.Init.InitChain(genesisBlock)
		return nil
	})
}

func (k *Kernel) AcceptUnpackedTxns() error {
	writings, err := k.subUnpackedWritings()
	if err != nil {
		return err
	}

	for _, txn := range writings {
		if k.CheckReplayAttack(txn) {
			continue
		}
		txn.FromP2P = true

		logrus.WithField("p2p", "accept-writing").
			Tracef("txn(%s) from network, content: %v", txn.TxnHash.String(), txn.Raw.WrCall)

		err = k.Pool.CheckTxn(txn)
		if err != nil {
			logrus.Error("check writing from P2P into txpool error: ", err)
			continue
		}
		err = k.Pool.Insert(txn)
		if err != nil {
			logrus.Error("insert writing from P2P into txpool error: ", err)
		}
	}

	topicTxnMap, err := k.subTopicWritings()
	if err != nil {
		return err
	}

	for topic, txns := range topicTxnMap {
		for _, txn := range txns {
			if txn == nil {
				continue
			}
			if k.CheckReplayAttack(txn) {
				continue
			}
			txn.FromP2P = true

			logrus.WithField("p2p", "accept-topic-writing").
				WithField("topic", topic).
				Tracef("txn(%s) from network, content: %v", txn.TxnHash.String(), txn.Raw.WrCall)

			err = k.Pool.CheckTxn(txn)
			if err != nil {
				logrus.WithError(err).WithField("topic", topic).Error("check topic writing from P2P into txpool error")
				continue
			}
			err = k.Pool.InsertWithTopic(topic, txn)
			if err != nil {
				logrus.WithError(err).WithField("topic", topic).Error("insert topic writing from P2P into txpool error")
			}
		}
	}

	return nil
}

func (k *Kernel) subUnpackedWritings() (types.SignedTxns, error) {
	byt, err := k.P2pNetwork.SubP2P(common.UnpackedWritingTopic)
	if err != nil {
		return nil, err
	}
	return types.DecodeSignedTxns(byt)
}

func (k *Kernel) pubUnpackedWritings(txns types.SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return k.P2pNetwork.PubP2P(common.UnpackedWritingTopic, byt)
}

func (k *Kernel) subTopicWritings() (map[string]types.SignedTxns, error) {
	if k.Land == nil {
		return nil, nil
	}
	topicTxns := make(map[string]types.SignedTxns)
	for _, topicTripod := range k.Land.OrderedTopicTripods() {
		if topicTripod.Topic == "" {
			continue
		}
		p2pTopic := common.TopicWritingTopic(topicTripod.Topic)
		byt, err := k.P2pNetwork.SubP2P(p2pTopic)
		if err != nil {
			if err == yerror.NoP2PTopic {
				continue
			}
			return nil, err
		}
		txns, err := types.DecodeSignedTxns(byt)
		if err != nil {
			return nil, err
		}
		if len(txns) == 0 {
			continue
		}
		topicTxns[p2pTopic] = txns
	}
	return topicTxns, nil
}

func (k *Kernel) pubTopicWritings(topic string, txns types.SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return k.P2pNetwork.PubP2P(topic, byt)
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
