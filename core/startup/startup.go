package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/synchronizer"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/env"
	"github.com/yu-org/yu/core/kernel"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/subscribe"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/txdb"
	"github.com/yu-org/yu/core/txpool"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/p2p"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/utils/codec"
	"path"
)

var (
	KernelCfg = new(config.KernelConf)

	Chain   types.IBlockChain
	TxnDB   types.ItxDB
	Pool    txpool.ItxPool
	StateDB state.IState

	Land = tripod.NewLand()
)

func DefaultStartup(tripodInstances ...interface{}) {
	k := InitDefaultKernel(tripodInstances...)
	k.Startup()
}

func StartUp(tripodInstances ...interface{}) {
	k := InitKernel(tripodInstances...)
	k.Startup()
}

func InitDefaultKernel(tripodInstances ...interface{}) *kernel.Kernel {
	tripodInstances = append([]interface{}{synchronizer.NewSynchronizer(KernelCfg.SyncMode)}, tripodInstances...)
	return InitKernel(tripodInstances...)
}

func InitKernel(tripodInstances ...interface{}) *kernel.Kernel {
	tripods := make([]*tripod.Tripod, 0)
	for _, v := range tripodInstances {
		tripods = append(tripods, tripod.ResolveTripod(v))
	}

	codec.GlobalCodec = &codec.RlpCodec{}
	gin.SetMode(gin.DebugMode)

	// init database
	KernelCfg.KVDB.Path = path.Join(KernelCfg.DataDir, KernelCfg.KVDB.Path)
	KernelCfg.BlockChain.ChainDB.Dsn = path.Join(KernelCfg.DataDir, KernelCfg.BlockChain.ChainDB.Dsn)
	kvdb, err := kv.NewKvdb(&KernelCfg.KVDB)
	if err != nil {
		logrus.Fatal("init kvdb error: ", err)
	}

	if TxnDB == nil {
		TxnDB = txdb.NewTxDB(KernelCfg.NodeType, kvdb)
	}
	if Chain == nil {
		Chain = blockchain.NewBlockChain(KernelCfg.NodeType, &KernelCfg.BlockChain, TxnDB)
	}
	if Pool == nil {
		Pool = txpool.WithDefaultChecks(KernelCfg.NodeType, &KernelCfg.Txpool, TxnDB)
	}
	if StateDB == nil {
		StateDB = state.NewStateDB(kvdb)
	}

	StartGrpcServer()

	for _, tri := range tripods {
		Pool.WithTripodCheck(tri)
	}

	chainEnv := &env.ChainEnv{
		State:      StateDB,
		Chain:      Chain,
		TxDB:       TxnDB,
		Pool:       Pool,
		Sub:        subscribe.NewSubscription(),
		P2pNetwork: p2p.NewP2P(&KernelCfg.P2P),
	}

	for i, t := range tripods {
		t.SetChainEnv(chainEnv)
		t.SetLand(Land)
		t.SetInstance(tripodInstances[i])
	}

	Land.SetTripods(tripods...)

	for _, tripodInterface := range tripodInstances {
		err = tripod.Inject(tripodInterface)
		if err != nil {
			logrus.Fatal("inject tripod failed: ", err)
		}
	}

	return kernel.NewKernel(KernelCfg, chainEnv, Land)
}
