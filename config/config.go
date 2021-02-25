package config

import (
	"github.com/BurntSushi/toml"
	. "yu/common"
)

type MasterConf struct {
	// 0: local-node
	// 1: master-worker
	RunMode RunMode `toml:"run_mode"`
	// serve http port
	HttpPort string `toml:"http_port"`
	// serve websocket port
	WsPort string `toml:"ws_port"`
	DB     KVconf `toml:"db"`
	// when beyond 'Timeout', it means this nodekeeper is down.
	Timeout int `toml:"timeout"`

	//---------P2P config--------
	// For listening from blockchain network.
	P2pListenAddrs []string `toml:"p2p_listen_addrs"`
	// To connect other hosts as a p2p network.
	ConnectAddrs []string `toml:"connect_addrs"`

	ProtocolID string `toml:"protocol_id"`
	// 0: RSA
	// 1: Ed25519
	// 2: Secp256k1
	// 3: ECDSA
	NodeKeyType int `toml:"node_key_type"`

	NodeKey string `toml:"node_key"`

	// Only RSA has this param.
	NodeKeyBits int `toml:"node_key_bits"`
	// When use param 'NodeKey', 'NodeKeyFile' will not work.
	NodeKeyFile string `toml:"node_key_file"`
}

type WorkerConf struct {
	Name           string `toml:"name"`
	DB             KVconf `toml:"db"`
	NodeKeeperPort string `toml:"node_keeper_port"`

	// serve http port
	HttpPort string `toml:"http_port"`
	// serve websocket port
	WsPort string `toml:"ws_port"`
	// the interval of heartbeat to NodeKeeper,
	// the unit is Second
	Interval int `toml:"interval"`

	TxPoolSize int `toml:"tx_pool_size"`
}

type NodeKeeperConf struct {
	ServesPort string `toml:"serves_port"`
	// Direction used to keep executable file and others.
	Dir string `toml:"dir"`
	// It MUST be {Dir}/xx.db
	// When you use {Dir}/path/to/xx.db, it will be trimmed as {Dir}/xx.db
	RepoDbPath   string `toml:"repo_db_path"`
	WorkerDbPath string `toml:"worker_db_path"`
	// specify the os and arch of repo
	// Usually you need not define it, it will get os and arch from local host.
	// such as: linux-amd64, darwin-amd64, windows-amd64, wasm
	OsArch string `toml:"os_arch"`

	MasterAddr   string `toml:"master_addr"`
	HeartbeatGap int    `toml:"heartbeat_gap"`
}

type TxpoolConf struct {
	PoolSize        uint64 `toml:"pool_size"`
	TxnMaxSize      int    `toml:"txn_max_size"`
	WaitTxnsTimeout int    `toml:"wait_txns_timeout"`
	DB              KVconf `toml:"db"`
}

type KVconf struct {
	// "bolt" "badger" "tikv"
	KvType string `toml:"kv_type"`
	// dbpath, such as boltdb, pebble
	Path string `toml:"path"`
	// distributed kvdb
	Hosts []string `toml:"hosts"`
}

func LoadConf(fpath string, cfg interface{}) (err error) {
	_, err = toml.DecodeFile(fpath, cfg)
	return
}
