package startup

import (
	"flag"
	"github.com/Lawliet-Chan/yu/blockchain"
	"github.com/Lawliet-Chan/yu/config"
	"github.com/Lawliet-Chan/yu/node/master"
	"github.com/Lawliet-Chan/yu/tripod"
	"github.com/Lawliet-Chan/yu/txpool"
	"github.com/Lawliet-Chan/yu/utils/codec"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	initLog()

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
		TxnsDB: config.SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "./txpool.db",
		},
		WorkerIP: "",
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
