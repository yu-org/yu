package keypair

import (
	"yu/common"
)

type KeyPair interface {

	Address() common.Address

	SignData([]byte) ([]byte, error)

	VerifySigner(msg, sig []byte) bool
}

func GenKeyPair(keyType string) (KeyPair, error) {
	switch keyType {
	case "sr25519":
		return genSr25519(), nil
	default:
		return genEd25519(), nil
	}
}