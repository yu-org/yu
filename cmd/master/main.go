package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"yu/apps/pow"
	"yu/blockchain"
	"yu/common"
	"yu/config"
	"yu/node/master"
	"yu/tripod"
	"yu/txpool"
)

func main() {
	var (
		masterCfgPath string
		chainCfgPath  string
		baseCfgPath   string
		txpoolCfgPath string

		masterCfg config.MasterConf
		chainCfg  config.BlockchainConf
		baseCfg   config.BlockBaseConf
		txpoolCfg config.TxpoolConf
	)

	flag.StringVar(&masterCfgPath, "m", "yu_conf/master.toml", "Master config file path")
	config.LoadConf(masterCfgPath, &masterCfg)

	flag.StringVar(&chainCfgPath, "c", "yu_conf/blockchain.toml", "blockchain config file path")
	config.LoadConf(chainCfgPath, &chainCfg)

	flag.StringVar(&baseCfgPath, "b", "yu_conf/blockbase.toml", "blockbase config file path")
	config.LoadConf(baseCfgPath, &baseCfg)

	flag.StringVar(&txpoolCfgPath, "tp", "yu_conf/txpool.toml", "txpool config file path")
	config.LoadConf(txpoolCfgPath, &txpoolCfg)

	initLog()

	chain, err := blockchain.NewBlockChain(&chainCfg)
	if err != nil {
		logrus.Panicf("load blockchain error: %s", err.Error())
	}
	base, err := blockchain.NewBlockBase(&baseCfg)
	if err != nil {
		logrus.Panicf("load blockbase error: %s", err.Error())
	}
	land := loadLand()

	var pool txpool.ItxPool
	switch masterCfg.RunMode {
	case common.LocalNode:
		pool = txpool.LocalWithDefaultChecks(&txpoolCfg)
	case common.MasterWorker:
		pool = txpool.ServerWithDefaultChecks(&txpoolCfg)
	}

	gin.SetMode(gin.ReleaseMode)

	m, err := master.NewMaster(&masterCfg, chain, base, pool, land)
	if err != nil {
		logrus.Panicf("load master error: %s", err.Error())
	}

	m.Startup()

}

func loadLand() *tripod.Land {
	land := tripod.NewLand()
	powTripod := pow.NewPow(1024)
	land.SetTripods(powTripod)
	return land
}

func initLog() {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logrus.InfoLevel)
}
