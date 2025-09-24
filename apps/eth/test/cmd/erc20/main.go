package main

import (
	"context"
	"flag"
	"github.com/yu-org/yu/apps/eth"
	"github.com/yu-org/yu/apps/eth/test/testx"
	"log"
	"os"
	"time"

	"github.com/yu-org/yu/apps/eth/config"
	"github.com/yu-org/yu/apps/eth/test/conf"
	"github.com/yu-org/yu/apps/eth/test/erc20"
)

var (
	evmConfigPath     string
	yuConfigPath      string
	poaConfigPath     string
	isParallel        bool
	nodeUrl           string
	genesisPrivateKey string
)

func init() {
	flag.StringVar(&evmConfigPath, "evmConfigPath", "./conf/eth.toml", "")
	flag.StringVar(&yuConfigPath, "yuConfigPath", "./conf/yu.toml", "")
	flag.StringVar(&poaConfigPath, "poaConfigPath", "./conf/poa.toml", "")
	flag.BoolVar(&isParallel, "parallel", true, "")
	flag.StringVar(&nodeUrl, "nodeUrl", "http://localhost:9092", "")
	flag.StringVar(&genesisPrivateKey, "key", "32e3b56c9f2763d2332e6e4188e4755815ac96441e899de121969845e343c2ff", "")

}

func main() {
	flag.Parse()
	yuCfg, poaCfg, evmConfig := testx.GenerateConfig(yuConfigPath, evmConfigPath, poaConfigPath)
	go func() {
		eth.StartupEthChain(yuCfg, poaCfg, evmConfig)
	}()
	time.Sleep(5 * time.Second)
	log.Println("finish start eth")
	if err := assertErc20Transfer(context.Background(), evmConfig); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("assert success")
	os.Exit(0)
}

func assertErc20Transfer(ctx context.Context, evmCfg *config.GethConfig) error {
	log.Println("start asserting transfer eth")
	ethManager := &erc20.EthManager{}
	cfg := conf.Config.EthCaseConf
	ethManager.Configure(cfg, nodeUrl, genesisPrivateKey, evmCfg.ChainConfig.ChainID.Int64())
	ethManager.AddTestCase(
		erc20.NewRandomTest("[rand_test 2 account, 1 transfer]", nodeUrl, 2, cfg.InitialEthCount, 1, evmCfg.ChainID),
	)
	return ethManager.Run(ctx)
}
