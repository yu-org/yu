package config

type Conf struct {
	NodeConf NodeConf `toml:"node_conf"`

	NodeDB KVconf `toml:"node_db"`

	NodeKeeperConf NodeKeeperConf `toml:"node_keeper_conf"`
}

type KVconf struct {
	// "bolt" "badger" "tikv"
	KvType string `toml:"kv_type"`
	// dbpath, such as boltdb, pebble
	Path string `toml:"path"`
	// distributed kvdb
	Hosts []string `toml:"hosts"`
}

type NodeConf struct {
	// 0: Master
	// 1: Worker
	NodeType uint   `toml:"node_type"`
	NodeName string `toml:"node_name"`

	// ------ Only Worker-Node has these params.
	NodeKeeperAddr   string `toml:"node_keeper_addr"`
	WorkerServesPort string `toml:"worker_serves_port"`

	// ------ Only Master-Node has these params.
	WorkersAddrs     []string `toml:"workers_addrs"`
	MasterServesPort string   `toml:"master_serves_port"`

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

type NodeKeeperConf struct {
	ServesPort string `toml:"serves_port"`
	// Direction used to keep executable file and others.
	Dir string `toml:"dir"`
	// It MUST be {Dir}/xx.db
	// When you use {Dir}/path/to/xx.db, it will be trimmed as {Dir}/xx.db
	RepoDbPath string `toml:"repo_db_path"`
	// specify the os and arch of repo
	// Usually you need not define it, it will get os and arch from local host.
	// such as: linux-amd64, darwin-amd64, windows-amd64, wasm
	OsArch string `toml:"os_arch"`

	MasterAddr string `toml:"master_addr"`
}
