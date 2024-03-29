package config

func InitDefaultCfg() *KernelConf {
	cfg := &KernelConf{
		RunMode:   0,
		DataDir:   "yu",
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
		KvType: "pebble",
		Path:   "yu.db",
		Hosts:  nil,
	}
	cfg.BlockChain = BlockchainConf{
		ChainDB: SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "chain.db",
		},
	}
	cfg.Txpool = TxpoolConf{
		PoolSize:   2048,
		TxnMaxSize: 1024000,
	}
	return cfg
}
