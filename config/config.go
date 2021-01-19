package config

type Conf struct {
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

	NodeType uint
}
