package startup

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/blockbase"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/chain_env"
	"github.com/yu-org/yu/core/kernel"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/subscribe"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/txpool"
	"github.com/yu-org/yu/infra/p2p"
	"github.com/yu-org/yu/utils/codec"
)

var (
	kernelCfgPath string
	kernelCfg     config.KernelConf
)

func StartUp(tripods ...tripod.Tripod) {
	initCfgFromFlags()
	initLog(kernelCfg.LogLevel)

	codec.GlobalCodec = &codec.RlpCodec{}
	gin.SetMode(gin.ReleaseMode)

	land := tripod.NewLand()
	land.SetTripods(tripods...)

	chain := blockchain.NewBlockChain(&kernelCfg.BlockChain)

	base := blockbase.NewBlockBase(&kernelCfg.BlockBase)

	statedb := state.NewStateDB(&kernelCfg.State)

	env := &chain_env.ChainEnv{
		IState:     statedb,
		Chain:      chain,
		Base:       base,
		Pool:       txpool.LocalWithDefaultChecks(&kernelCfg.Txpool),
		Sub:        subscribe.NewSubscription(),
		P2pNetwork: p2p.NewP2P(&kernelCfg.P2P),
	}

	for _, t := range tripods {
		t.SetChainEnv(env)
	}

	m := kernel.NewKernel(&kernelCfg, env, land)

	m.Startup()
}

func initCfgFromFlags() {
	useDefaultCfg := flag.Bool("dc", false, "default config files")

	flag.StringVar(&kernelCfgPath, "k", "yu_conf/kernel.toml", "Kernel config file path")

	flag.Parse()
	if *useDefaultCfg {
		initDefaultCfg()
		return
	}

	config.LoadConf(kernelCfgPath, &kernelCfg)
}

func initLog(level string) {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		panic("parse log level error: " + err.Error())
	}

	logrus.SetLevel(lvl)
}

func initDefaultCfg() {
	kernelCfg = config.KernelConf{
		RunMode:  0,
		HttpPort: "7999",
		WsPort:   "8999",
		LogLevel: "info",
		LeiLimit: 50000,
		NkDB: config.KVconf{
			KvType: "bolt",
			Path:   "./nk_db.db",
			Hosts:  nil,
		},
		Timeout: 60,
	}
	kernelCfg.P2P = config.P2pConf{
		P2pListenAddrs:  []string{"/ip4/127.0.0.1/tcp/8887"},
		Bootnodes:       nil,
		ProtocolID:      "yu",
		NodeKeyType:     1,
		NodeKeyRandSeed: 1,
		NodeKey:         "",
		NodeKeyBits:     0,
		NodeKeyFile:     "",
	}
	kernelCfg.BlockChain = config.BlockchainConf{
		ChainDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "chain.db",
		},
		BlocksFromP2pDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "blocks_from_p2p.db",
		},
	}
	kernelCfg.BlockBase = config.BlockBaseConf{
		BaseDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "blockbase.db",
		}}
	kernelCfg.Txpool = config.TxpoolConf{
		PoolSize:   2048,
		TxnMaxSize: 1024000,
		DB: config.KVconf{
			KvType: "badger",
			Path:   "txpool.db",
			Hosts:  nil,
		},
		WorkerIP: "",
	}
	kernelCfg.State = config.StateConf{KV: config.StateKvConf{
		IndexDB: config.KVconf{
			KvType: "bolt",
			Path:   "./state_index.db",
			Hosts:  nil,
		},
		NodeBase: config.KVconf{
			KvType: "bolt",
			Path:   "./state_base.db",
			Hosts:  nil,
		},
	}}
}
