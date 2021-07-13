package startup

import (
	"flag"
	"github.com/Lawliet-Chan/yu/blockchain"
	"github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/config"
	"github.com/Lawliet-Chan/yu/node/master"
	"github.com/Lawliet-Chan/yu/state"
	"github.com/Lawliet-Chan/yu/tripod"
	"github.com/Lawliet-Chan/yu/txpool"
	"github.com/Lawliet-Chan/yu/utils/codec"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	masterCfgPath string
	chainCfgPath  string
	baseCfgPath   string
	txpoolCfgPath string
	stateCfgPath  string

	masterCfg config.MasterConf
	chainCfg  config.BlockchainConf
	baseCfg   config.BlockBaseConf
	txpoolCfg config.TxpoolConf
	stateCfg  config.StateConf
)

func StartUp(tripods ...tripod.Tripod) {
	initCfgFromFlags()
	initLog()

	codec.GlobalCodec = &codec.RlpCodec{}
	gin.SetMode(gin.ReleaseMode)

	chain, err := blockchain.NewBlockChain(&chainCfg)
	if err != nil {
		logrus.Panicf("load blockchain error: %s", err.Error())
	}
	base, err := blockchain.NewBlockBase(&baseCfg)
	if err != nil {
		logrus.Panicf("load blockbase error: %s", err.Error())
	}

	var pool txpool.ItxPool
	switch masterCfg.RunMode {
	case common.LocalNode:
		pool = txpool.LocalWithDefaultChecks(&txpoolCfg)
	case common.MasterWorker:
		logrus.Panic("no server txpool")
	}

	stateStore, err := state.NewStateStore(&stateCfg)
	if err != nil {
		logrus.Panicf("load stateKV error: %s", err.Error())
	}

	land := tripod.NewLand()
	land.SetTripods(tripods...)

	m, err := master.NewMaster(&masterCfg, chain, base, pool, stateStore, land)
	if err != nil {
		logrus.Panicf("load master error: %s", err.Error())
	}

	m.Startup()
}

func initCfgFromFlags() {
	useDefaultCfg := flag.Bool("dc", false, "default config files")
	if *useDefaultCfg {
		initDefaultCfg()
		return
	}

	flag.StringVar(&masterCfgPath, "m", "yu_conf/master.toml", "Master config file path")
	config.LoadConf(masterCfgPath, &masterCfg)

	flag.StringVar(&chainCfgPath, "c", "yu_conf/blockchain.toml", "blockchain config file path")
	config.LoadConf(chainCfgPath, &chainCfg)

	flag.StringVar(&baseCfgPath, "b", "yu_conf/blockbase.toml", "blockbase config file path")
	config.LoadConf(baseCfgPath, &baseCfg)

	flag.StringVar(&txpoolCfgPath, "tp", "yu_conf/txpool.toml", "txpool config file path")
	config.LoadConf(txpoolCfgPath, &txpoolCfg)

	flag.StringVar(&stateCfgPath, "s", "yu_conf/state.toml", "state config file path")
	config.LoadConf(stateCfgPath, &stateCfg)
}

func initLog() {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logrus.InfoLevel)
}

func initDefaultCfg() {
	masterCfg = config.MasterConf{
		RunMode:  0,
		HttpPort: "7999",
		WsPort:   "8999",
		NkDB: config.KVconf{
			KvType: "bolt",
			Path:   "./nk_db.db",
			Hosts:  nil,
		},
		Timeout:         60,
		P2pListenAddrs:  []string{"/ip4/127.0.0.1/tcp/8887"},
		ConnectAddrs:    nil,
		ProtocolID:      "yu",
		NodeKeyType:     1,
		NodeKeyRandSeed: 1,
		NodeKey:         "",
		NodeKeyBits:     0,
		NodeKeyFile:     "",
	}
	chainCfg = config.BlockchainConf{
		ChainDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "chain.db",
		},
		BlocksFromP2pDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "blocks_from_p2p.db",
		},
	}
	txpoolCfg = config.TxpoolConf{
		PoolSize:   2048,
		TxnMaxSize: 1024000,
		Timeout:    10,
		TxnsDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "./txpool.db",
		},
		WorkerIP: "",
	}
	stateCfg = config.StateConf{KV: config.StateKvConf{
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
