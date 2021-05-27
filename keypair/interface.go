package keypair

import (
	"github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/yerror"
)

const (
	Sr25519 = "sr25519"
	Ed25519 = "ed25519"
)

func GenKeyPair(keyType string) (PubKey, PrivKey, error) {
	switch keyType {
	case Sr25519:
		pub, priv := genSr25519()
		return pub, priv, nil
	case Ed25519:
		pub, priv := genEd25519()
		return pub, priv, nil
	default:
		return nil, nil, NoKeyType
	}
}

func PubKeyFromBytes(keyType string, data []byte) (PubKey, error) {
	switch keyType {
	case Sr25519:
		return srPubKeyFromBytes(data), nil
	case Ed25519:
		return edPubKeyFromBytes(data), nil
	default:
		return nil, NoKeyType
	}
}

type Key interface {
	Type() string
	Equals(key Key) bool
	Bytes() []byte
	String() string
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
