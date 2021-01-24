package config

type Conf struct {
	NodeConf NodeConf

	NodeDB KVconf

	NodeKeeperConf NodeKeeperConf
}

type KVconf struct {
	// "bolt" "badger" "tikv"
	KVtype string
	// dbpath, such as boltdb, pebble
	Path string
	// distributed kvdb
	Hosts []string
}

type NodeConf struct {
	// 0: Master
	// 1: Worker
	NodeType uint
	NodeName string

	// ------ Only Worker-Node has these params.
	MasterNode       string
	WorkerServesPort string

	// ------ Only Master-Node has these params.
	WorkerNodes      []string
	MasterServesPort string

	//---------P2P config--------
	// For listening from blockchain network.
	P2pListenAddrs []string
	// To connect other hosts as a p2p network.
	ConnectAddrs []string

	ProtocolID string
	// 0: RSA
	// 1: Ed25519
	// 2: Secp256k1
	// 3: ECDSA
	NodeKeyType int

	NodeKey string

	// Only RSA has this param.
	NodeKeyBits int
	// When use param 'NodeKey', 'NodeKeyFile' will not work.
	NodeKeyFile string
}

type NodeKeeperConf struct {
	ServesPort string
	// Direction used to keep executable file
	BinaryDir string
}
