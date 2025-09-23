package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"time"

	"golang.org/x/time/rate"

	"github.com/yu-org/yu/apps/eth/config"
	"github.com/yu-org/yu/apps/eth/test/conf"
	"github.com/yu-org/yu/apps/eth/test/pkg"
	"github.com/yu-org/yu/apps/eth/test/transfer"
)

var (
	configPath        string
	evmConfigPath     string
	qps               int
	duration          time.Duration
	action            string
	preCreateWallets  int
	nodeUrl           string
	genesisPrivateKey string
)

const benchmarkDataPath = "./bin/eth_benchmark_data.json"

func init() {
	flag.StringVar(&configPath, "configPath", "", "")
	flag.StringVar(&evmConfigPath, "evmConfigPath", "./conf/evm.toml", "")
	flag.IntVar(&qps, "qps", 10000, "")
	flag.DurationVar(&duration, "duration", 5*time.Minute, "")
	flag.StringVar(&action, "action", "run", "")
	flag.IntVar(&preCreateWallets, "preCreateWallets", 100, "")
	flag.StringVar(&nodeUrl, "nodeUrl", "http://localhost:9092", "")
	flag.StringVar(&genesisPrivateKey, "key", "32e3b56c9f2763d2332e6e4188e4755815ac96441e899de121969845e343c2ff", "")

}

func main() {
	flag.Parse()
	if err := conf.LoadConfig(configPath); err != nil {
		panic(err)
	}
	evmConfig := config.LoadGethConfig(evmConfigPath)
	switch action {
	case "prepare":
		prepareBenchmark(evmConfig.ChainConfig.ChainID.Int64())
	case "run":
		blockBenchmark(evmConfig.ChainConfig.ChainID.Int64(), qps)
	}
}

func prepareBenchmark(chainID int64) error {
	ethManager := &transfer.EthManager{}
	cfg := conf.Config.EthCaseConf
	ethManager.Configure(cfg, nodeUrl, genesisPrivateKey, chainID)
	wallets, err := ethManager.PreCreateWallets(preCreateWallets, cfg.InitialEthCount)
	if err != nil {
		return err
	}
	_, err = os.Stat(benchmarkDataPath)
	if err == nil {
		os.Remove(benchmarkDataPath)
	}
	file, err := os.Create(benchmarkDataPath)
	if err != nil {
		return err
	}
	defer file.Close()
	d, err := json.Marshal(wallets)
	if err != nil {
		return err
	}
	_, err = file.Write(d)
	return err
}

func loadWallets() ([]*pkg.EthWallet, error) {
	d, err := os.ReadFile(benchmarkDataPath)
	if err != nil {
		return nil, err
	}
	exp := make([]*pkg.EthWallet, 0)
	if err := json.Unmarshal(d, &exp); err != nil {
		return nil, err
	}
	return exp, nil
}

func blockBenchmark(chainID int64, qps int) error {
	wallets, err := loadWallets()
	if err != nil {
		return err
	}
	ethManager := &transfer.EthManager{}
	cfg := conf.Config.EthCaseConf
	ethManager.Configure(cfg, nodeUrl, genesisPrivateKey, chainID)
	limiter := rate.NewLimiter(rate.Limit(qps), qps)
	ethManager.AddTestCase(transfer.NewRandomBenchmarkTest("[rand_test 1000 transfer]", cfg.InitialEthCount, wallets, limiter))
	runBenchmark(ethManager)
	return nil
}

func runBenchmark(manager *transfer.EthManager) {
	after := time.After(duration)
	for {
		select {
		case <-after:
			return
		default:
		}
		manager.Run(context.Background())
	}
}
