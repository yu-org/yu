package startup

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yu-altar/yu/blockchain"
	"github.com/yu-altar/yu/config"
	"github.com/yu-altar/yu/node/master"
	"github.com/yu-altar/yu/tripod"
	"github.com/yu-altar/yu/txpool"
	"github.com/yu-altar/yu/utils/codec"
)

var (
	masterCfgPath string
	masterCfg     config.MasterConf
)

var (
	Chain  blockchain.IBlockChain
	Base   blockchain.IBlockBase
	TxPool txpool.ItxPool
)

func StartUp(tripods ...tripod.Tripod) {
	initCfgFromFlags()
	initLog(masterCfg.LogLevel)

	codec.GlobalCodec = &codec.RlpCodec{}
	gin.SetMode(gin.ReleaseMode)

	land := tripod.NewLand()
	land.SetTripods(tripods...)

	m, err := master.NewMaster(&masterCfg, Chain, Base, TxPool, land)
	if err != nil {
		logrus.Panicf("load master error: %s", err.Error())
	}

	m.Startup()
}

func initCfgFromFlags() {
	useDefaultCfg := flag.Bool("dc", false, "default config files")

	flag.StringVar(&masterCfgPath, "m", "yu_conf/master.toml", "Master config file path")

	flag.Parse()
	if *useDefaultCfg {
		initDefaultCfg()
		return
	}

	config.LoadConf(masterCfgPath, &masterCfg)
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
	masterCfg = config.MasterConf{
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
		Timeout:         60,
		P2pListenAddrs:  []string{"/ip4/127.0.0.1/tcp/8887"},
		Bootnodes:       nil,
		ProtocolID:      "yu",
		NodeKeyType:     1,
		NodeKeyRandSeed: 1,
		NodeKey:         "",
		NodeKeyBits:     0,
		NodeKeyFile:     "",
	}
	masterCfg.BlockChain = config.BlockchainConf{
		ChainDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "chain.db",
		},
		BlocksFromP2pDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "blocks_from_p2p.db",
		},
	}
	masterCfg.BlockBase = config.BlockBaseConf{
		BaseDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "blockbase.db",
		}}
	masterCfg.Txpool = config.TxpoolConf{
		PoolSize:   2048,
		TxnMaxSize: 1024000,
		Timeout:    10,
		WorkerIP:   "",
	}
	masterCfg.State = config.StateConf{KV: config.StateKvConf{
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
