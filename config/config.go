package config

import (
	"github.com/BurntSushi/toml"
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/sirupsen/logrus"
)

type MasterConf struct {
	// 0: local-node
	// 1: master-worker
	RunMode RunMode `toml:"run_mode"`
	// serve http port
	HttpPort string `toml:"http_port"`
	// serve websocket port
	WsPort string `toml:"ws_port"`

	LeiLimit uint64 `toml:"lei_limit"`

	NkDB KVconf `toml:"nk_db"`
	// when beyond 'Timeout', it means this nodekeeper is down.
	// Unit is Second.
	Timeout int `toml:"timeout"`

	//---------component config---------
	BlockChain BlockchainConf `toml:"block_chain"`
	BlockBase  BlockBaseConf  `toml:"block_base"`
	State      StateConf      `toml:"state"`
	Txpool     TxpoolConf     `toml:"txpool"`

	//---------P2P config--------
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

//type WorkerConf struct {
//	Name           string `toml:"name"`
//	DB             KVconf `toml:"db"`
//	NodeKeeperPort string `toml:"node_keeper_port"`
//
//	// number of broadcast txns every time
//	NumOfBcTxns int `toml:"num_of_bc_txns"`
//
//	// serve http port
//	HttpPort string `toml:"http_port"`
//	// serve websocket port
//	WsPort string `toml:"ws_port"`
//	// the interval of heartbeat to NodeKeeper,
//	// the unit is Second
//	Interval int `toml:"interval"`
//
//	TxPoolSize int `toml:"tx_pool_size"`
//}
//
//type NodeKeeperConf struct {
//	ServesPort string `toml:"serves_port"`
//	// Direction used to keep executable file and others.
//	Dir string `toml:"dir"`
//	// It MUST be {Dir}/xx.db
//	// When you use {Dir}/path/to/xx.db, it will be trimmed as {Dir}/xx.db
//	RepoDbPath   string `toml:"repo_db_path"`
//	WorkerDbPath string `toml:"worker_db_path"`
//	// specify the os and arch of repo
//	// Usually you need not define it, it will get os and arch from local host.
//	// such as: linux-amd64, darwin-amd64, windows-amd64, wasm
//	OsArch string `toml:"os_arch"`
//
//	MasterAddr   string `toml:"master_addr"`
//	HeartbeatGap int    `toml:"heartbeat_gap"`
//}

type BlockchainConf struct {
	ChainDB         SqlDbConf `toml:"chain_db"`
	BlocksFromP2pDB SqlDbConf `toml:"blocks_from_p2p_db"`
}

type BlockBaseConf struct {
	BaseDB SqlDbConf `toml:"base_db"`
}

type TxpoolConf struct {
	PoolSize   uint64    `toml:"pool_size"`
	TxnMaxSize int       `toml:"txn_max_size"`
	Timeout    int       `toml:"timeout"`
	TxnsDB     SqlDbConf `toml:"txns_db"`
	// In local-node mode, this will be null.
	WorkerIP string `toml:"worker_ip"`
}

func LoadConf(fpath string, cfg interface{}) {
	_, err := toml.DecodeFile(fpath, cfg)
	if err != nil {
		logrus.Panicf("load config-file(%s) error: %s ", fpath, err.Error())
	}
}
