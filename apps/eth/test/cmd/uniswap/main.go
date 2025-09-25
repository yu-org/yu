package main

import (
	"context"
	"flag"
	"os"
	"runtime"
	"time"

	"github.com/yu-org/yu/apps/eth"

	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/apps/eth/test/conf"
	"github.com/yu-org/yu/apps/eth/test/testx"
	"github.com/yu-org/yu/apps/eth/test/transfer"
	"github.com/yu-org/yu/apps/eth/test/uniswap"
)

var (
	evmConfigPath     string
	yuConfigPath      string
	poaConfigPath     string
	nodeUrl           string
	genesisPrivateKey string
)

func init() {
	flag.StringVar(&evmConfigPath, "evmConfigPath", "./conf/eth.toml", "")
	flag.StringVar(&yuConfigPath, "yuConfigPath", "./conf/yu.toml", "")
	flag.StringVar(&poaConfigPath, "poaConfigPath", "./conf/poa.toml", "")
	flag.StringVar(&nodeUrl, "nodeUrl", "http://localhost:9092", "")
	flag.StringVar(&genesisPrivateKey, "key", "32e3b56c9f2763d2332e6e4188e4755815ac96441e899de121969845e343c2ff", "")

}

func main() {
	flag.Parse()
	yuCfg, poaCfg, evmConfig := testx.GenerateConfig(yuConfigPath, evmConfigPath, poaConfigPath)
	go func() {
		logrus.Infof("Number of goroutines after app.Start: %d", runtime.NumGoroutine())
		eth.StartupEthChain(yuCfg, poaCfg, evmConfig)
	}()
	time.Sleep(5 * time.Second)
	logrus.Info("finish start eth")
	if err := assertUniswapV2(context.Background(), evmConfig.ChainConfig.ChainID.Int64()); err != nil {
		logrus.Info(err)
		os.Exit(1)
	}
	logrus.Info("assert success")
	os.Exit(0)
}

func assertUniswapV2(ctx context.Context, chainID int64) error {
	ethManager := &transfer.EthManager{}
	cfg := conf.Config.EthCaseConf
	ethManager.Configure(cfg, nodeUrl, genesisPrivateKey, chainID)
	ethManager.AddTestCase(
		uniswap.NewUniswapV2AccuracyTestCase("UniswapV2 Accuracy TestCase", nodeUrl, 2, cfg.InitialEthCount, chainID),
	)
	return ethManager.Run(ctx)
}
