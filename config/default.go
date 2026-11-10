package config

import "path"

// InitDefaultCfg returns defaults for a new chain using SQLite table state.
func InitDefaultCfg() *KernelConf {
	dataDir := "yu"
	cfg := &KernelConf{
		ChainSpec:   DefaultChainSpec(),
		StatedbType: "sql",
		StateSQL:    SqlDbConf{SqlDbType: "sqlite", Dsn: "state.db"},
		RunMode:     0,
		DataDir:     dataDir,
		HttpPort:    "7999",
		WsPort:      "8999",
		LogLevel:    "info",
		LogOutput:   path.Join(dataDir, "yu.log"),
		LeiLimit:    50000,
		EnablePProf: true,
		PProfPort:   "10199",
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
		ChainID: 0,
		ChainDB: SqlDbConf{
			SqlDbType: "sqlite",
			Dsn:       "chain.db",
		},
		CacheSize: 10,
	}
	cfg.Txpool = TxpoolConf{
		PoolSize:   2048,
		TxnMaxSize: 1024000,
	}
	return cfg
}
