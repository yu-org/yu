package config

type Conf struct {
	NodeConf NodeConf

	MasterNodeDB KVconf
}

type KVconf struct {
	// "bolt" "badger" "tikv"
	KVtype string
	// embedded kvdb, such as boltdb, pebble
	Path string
	// distributed kvdb
	Hosts []string
}

type NodeConf struct {
	// 0: Master
	// 1: Worker
	NodeType uint
	NodeName string
	// For listening from blockchain network.
	P2pAddrs []string
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
