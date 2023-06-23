package config

import (
	"os"
	"path"
)

func InitDefaultCfg() *KernelConf {
	return InitDefaultCfgWithDir("")
}

func InitDefaultCfgWithDir(dir string) *KernelConf {
	if dir != "" {
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			panic(err)
		}
	}

	cfg := &KernelConf{
		RunMode:   0,
		HttpPort:  "7999",
		WsPort:    "8999",
		LogLevel:  "info",
		LogOutput: "yu.log",
		LeiLimit:  50000,
	}
	cfg.P2P = P2pConf{
		P2pListenAddrs:  []string{"/ip4/127.0.0.1/tcp/8887"},
		Bootnodes:       nil,
		ProtocolID:      "yu",
		NodeKeyType:     1,
		NodeKeyRandSeed: 1,
		NodeKey:         "",
		NodeKeyBits:     0,
		NodeKeyFile:     "",
	}
	cfg.KVDB = KVconf{
		KvType: "bolt",
		Path:   path.Join(dir, "yu.db"),
		Hosts:  nil,
	}
	cfg.BlockChain = BlockchainConf{
		ChainDB: SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       path.Join(dir, "chain.db"),
		},
	}
	cfg.Txpool = TxpoolConf{
		PoolSize:   2048,
		TxnMaxSize: 1024000,
	}
	return cfg
}
