package keypair

import (
	"yu/common"
)

func GenKeyPair(keyType string) (PubKey, PrivKey) {
	switch keyType {
	case "sr25519":
		return genSr25519()
	default:
		return genEd25519()
	}
}

type Key interface {
	Type() string
	Equals(key Key) bool
	Bytes() []byte
}

type PubKey interface {
	Key
	Address() common.Address
	VerifySignature(msg, sig []byte) bool
}

type PrivKey interface {
	Key
	SignData([]byte) ([]byte, error)
}
