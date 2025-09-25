package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/yu-org/yu/apps/eth"

	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/apps/eth/test/conf"
	"github.com/yu-org/yu/apps/eth/test/testx"
	"github.com/yu-org/yu/apps/eth/test/transfer"
)

var (
	evmConfigPath     string
	yuConfigPath      string
	poaConfigPath     string
	nodeUrl           string
	genesisPrivateKey string
	AsClient          bool
	chainID           int64
)

func init() {
	flag.StringVar(&evmConfigPath, "evmConfigPath", "./conf/eth.toml", "")
	flag.StringVar(&yuConfigPath, "yuConfigPath", "./conf/yu.toml", "")
	flag.StringVar(&poaConfigPath, "poaConfigPath", "./conf/poa.toml", "")

	flag.StringVar(&nodeUrl, "nodeUrl", "http://localhost:9092", "")
	flag.StringVar(&genesisPrivateKey, "key", "32e3b56c9f2763d2332e6e4188e4755815ac96441e899de121969845e343c2ff", "")

	flag.BoolVar(&AsClient, "as-client", false, "")
	flag.Int64Var(&chainID, "chainId", 50341, "")

}

func main() {
	flag.Parse()
	if !AsClient {
		yuCfg, poaCfg, evmConfig := testx.GenerateConfig(yuConfigPath, evmConfigPath, poaConfigPath)
		go func() {
			eth.StartupEthChain(yuCfg, poaCfg, evmConfig)
		}()
		time.Sleep(5 * time.Second)
		logrus.Info("finish start eth")
		chainID = evmConfig.ChainConfig.ChainID.Int64()
	}
	if err := assertEthTransfer(context.Background(), chainID); err != nil {
		logrus.Info(err)
		os.Exit(1)
	}
	logrus.Info("assert success")
	os.Exit(0)
}

func assertEthTransfer(ctx context.Context, chainID int64) error {
	logrus.Info("start asserting transfer eth")
	ethManager := &transfer.EthManager{}
	cfg := conf.Config.EthCaseConf
	ethManager.Configure(cfg, nodeUrl, genesisPrivateKey, chainID)
	ethManager.AddTestCase(
		transfer.NewRandomTest("[rand_test 2 account, 1 transfer]", 2, cfg.InitialEthCount, 1),
		transfer.NewRandomTest("[rand_test 20 account, 100 transfer]", 20, cfg.InitialEthCount, 100),
		transfer.NewConflictTest("[conflict_test 20 account, 50 transfer]", 20, cfg.InitialEthCount, 50),
	)
	return ethManager.Run(ctx)
}
