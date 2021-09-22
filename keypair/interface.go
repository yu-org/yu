package keypair

import (
	"github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/yerror"
)

const (
	Sr25519 = "sr25519"
	Ed25519 = "ed25519"

	Sr25519Idx = "1"
	Ed25519Idx = "2"
)

var KeyTypeBytLen = 1

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

// data: (keyTypeBytes + keyBytes)
func PubKeyFromBytes(data []byte) (PubKey, error) {
	keyTypeByt := data[:KeyTypeBytLen]
	switch string(keyTypeByt) {
	case Sr25519Idx:
		return SrPubKeyFromBytes(data[KeyTypeBytLen:]), nil
	case Ed25519Idx:
		return EdPubKeyFromBytes(data[KeyTypeBytLen:]), nil
	default:
		return nil, NoKeyType
	}
}

func PubkeyFromStr(data string) (PubKey, error) {
	byt := common.FromHex(data)
	return PubKeyFromBytes(byt)
}

type Key interface {
	Type() string
	Equals(key Key) bool
	Bytes() []byte
	String() string

	BytesWithType() []byte
	StringWithType() string
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
