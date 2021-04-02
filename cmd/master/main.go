package main

import (
	"flag"
	"github.com/sirupsen/logrus"
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

	chain := blockchain.NewKvBlockChain(&chainCfg)
	base := blockchain.NewBlockBase(&baseCfg)
	land := tripod.NewLand()

	//todo: set tripods into land

	var pool txpool.ItxPool
	switch masterCfg.RunMode {
	case common.LocalNode:
		pool = txpool.LocalWithDefaultChecks(&txpoolCfg, land)
	case common.MasterWorker:
		pool = txpool.ServerWithDefaultChecks(&txpoolCfg, land)
	}

	m, err := master.NewMaster(&masterCfg, chain, base, pool, land)
	if err != nil {
		logrus.Panicf("load master error: %s", err.Error())
	}

	m.Startup()

}
