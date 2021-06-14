package main

import (
	"flag"
	"github.com/Lawliet-Chan/yu/apps"
	"github.com/Lawliet-Chan/yu/blockchain"
	"github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/config"
	"github.com/Lawliet-Chan/yu/node/master"
	"github.com/Lawliet-Chan/yu/state"
	"github.com/Lawliet-Chan/yu/txpool"
	"github.com/Lawliet-Chan/yu/utils/codec"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
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

	initLog()

	codec.GlobalCodec = &codec.RlpCodec{}

	chain, err := blockchain.NewBlockChain(&chainCfg)
	if err != nil {
		logrus.Panicf("load blockchain error: %s", err.Error())
	}
	base, err := blockchain.NewBlockBase(&baseCfg)
	if err != nil {
		logrus.Panicf("load blockbase error: %s", err.Error())
	}
	land := apps.LoadLand()

	var pool txpool.ItxPool
	switch masterCfg.RunMode {
	case common.LocalNode:
		pool = txpool.LocalWithDefaultChecks(&txpoolCfg)
	case common.MasterWorker:
		pool = txpool.ServerWithDefaultChecks(&txpoolCfg)
	}

	gin.SetMode(gin.ReleaseMode)

	stateStore, err := state.NewStateStore(&stateCfg)
	if err != nil {
		logrus.Panicf("load stateKV error: %s", err.Error())
	}

	m, err := master.NewMaster(&masterCfg, chain, base, pool, stateStore, land)
	if err != nil {
		logrus.Panicf("load master error: %s", err.Error())
	}

	m.Startup()

}

func initLog() {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logrus.InfoLevel)
}
