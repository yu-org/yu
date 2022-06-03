package config

import (
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
)

type KernelConf struct {
	// 0: local-node
	// 1: master-worker
	RunMode RunMode `toml:"run_mode"`
	// serve http port
	HttpPort string `toml:"http_port"`
	// serve websocket port
	WsPort string `toml:"ws_port"`
	// log out level:
	// panic, fatal, error, warn, info, debug, trace
	LogLevel string `toml:"log_level"`
	// log out put file path
	LogOutput string `toml:"log_output"`

	LeiLimit uint64 `toml:"lei_limit"`

	KVDB KVconf `toml:"kvdb"`
	//---------component config---------
	BlockChain BlockchainConf `toml:"block_chain"`
	YuDB       YuDBConf       `toml:"yu_db"`
	State      StateConf      `toml:"state"`
	Txpool     TxpoolConf     `toml:"txpool"`
	P2P        P2pConf        `toml:"p2p"`
}

type P2pConf struct {
	// For listening from blockchain network.
	P2pListenAddrs []string `toml:"p2p_listen_addrs"`
	// To connect other hosts as a p2p network.
	Bootnodes []string `toml:"bootnodes"`

	ProtocolID string `toml:"protocol_id"`
	// 0: RSA
	// 1: Ed25519
	// 2: Secp256k1
	// 3: ECDSA
	NodeKeyType int `toml:"node_key_type"`

	NodeKeyRandSeed int64 `toml:"node_key_rand_seed"`

	NodeKey string `toml:"node_key"`

	// Only RSA has this param.
	NodeKeyBits int `toml:"node_key_bits"`
	// When use param 'NodeKey', 'NodeKeyFile' will not work.
	NodeKeyFile string `toml:"node_key_file"`
}

type BlockchainConf struct {
	ChainDB SqlDbConf `toml:"chain_db"`
}

type YuDBConf struct {
	BaseDB SqlDbConf `toml:"base_db"`
}

type TxpoolConf struct {
	PoolSize   uint64 `toml:"pool_size"`
	TxnMaxSize int    `toml:"txn_max_size"`
}

func LoadConf(fpath string, cfg interface{}) {
	_, err := toml.DecodeFile(fpath, cfg)
	if err != nil {
		logrus.Panicf("load config-file(%s) error: %s ", fpath, err.Error())
	}
}
