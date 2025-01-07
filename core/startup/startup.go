package startup

import (
	"os"
	"path"

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
)

var (
	// KernelCfg = new(config.KernelConf)

	Chain   types.IBlockChain
	TxnDB   types.ItxDB
	Pool    txpool.ItxPool
	StateDB state.IState

	Land = tripod.NewLand()
)

func DefaultStartup(cfg *config.KernelConf) {
	k := InitDefaultKernel(cfg)
	k.Startup()
}

func StartUp(cfg *config.KernelConf) {
	k := InitKernel(cfg)
	k.Startup()
}

func InitDefaultKernel(cfg *config.KernelConf) *kernel.Kernel {
	return InitKernel(cfg).WithTripods(synchronizer.NewSynchronizer(cfg.SyncMode))
}

func InitKernel(cfg *config.KernelConf) *kernel.Kernel {
	beforeStartUp(cfg)

	codec.GlobalCodec = &codec.RlpCodec{}
	gin.SetMode(gin.ReleaseMode)

	// init database
	cfg.KVDB.Path = path.Join(cfg.DataDir, cfg.KVDB.Path)

	if cfg.BlockChain.ChainDB.Dsn == "" {
		cfg.BlockChain.ChainDB.Dsn = os.Getenv("chain_db_dsn")
	}
	if cfg.BlockChain.ChainDB.SqlDbType == "sqlite" {
		cfg.BlockChain.ChainDB.Dsn = path.Join(cfg.DataDir, cfg.BlockChain.ChainDB.Dsn)
	}

	kvdb, err := kv.NewKvdb(&cfg.KVDB)
	if err != nil {
		logrus.Fatal("init kvdb error: ", err)
	}

	if TxnDB == nil {
		TxnDB = txdb.NewTxDB(cfg.NodeType, kvdb)
	}
	if Chain == nil {
		Chain = blockchain.NewBlockChain(cfg.NodeType, &cfg.BlockChain, TxnDB)
	}
	if Pool == nil {
		Pool = txpool.WithDefaultChecks(cfg.NodeType, &cfg.Txpool)
	}
	if StateDB == nil {
		StateDB = state.NewStateDB(cfg.StatedbType, kvdb)
	}

	// StartGrpcServer(cfg)

	chainEnv := &env.ChainEnv{
		State:      StateDB,
		Chain:      Chain,
		TxDB:       TxnDB,
		Pool:       Pool,
		Sub:        subscribe.NewSubscription(),
		P2pNetwork: p2p.NewP2P(&cfg.P2P),
	}

	return kernel.NewKernel(cfg, chainEnv, Land)
}
