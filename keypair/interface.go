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

type PubKey interface {
	Address() common.Address

	VerifySignature(msg, sig []byte) bool
}

type PrivKey interface {
	SignData([]byte) ([]byte, error)
}