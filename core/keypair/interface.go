package keypair

import (
	"github.com/pkg/errors"
	"github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
)

const (
	SecretFree = "secret-free"
	Sr25519    = "sr25519"
	Ed25519    = "ed25519"
	Secp256k1  = "secp256k1"

	SecretFreeIdx = "0"
	Sr25519Idx    = "1"
	Ed25519Idx    = "2"
	Secp256k1Idx  = "3"
)

var KeyTypeBytLen = 1

func GenKeyPairWithSecret(keyType string, secret []byte) (PubKey, PrivKey, error) {
	switch keyType {
	case SecretFree:
		return nil, nil, nil
	case Sr25519:
		pub, priv := GenSrKeyWithSecret(secret)
		return pub, priv, nil
	case Ed25519:
		pub, priv := GenEdKeyWithSecret(secret)
		return pub, priv, nil
	case Secp256k1:
		pub, priv := GenSecpKeyWithSecret(secret)
		return pub, priv, nil
	default:
		return nil, nil, NoKeyType
	}
}

func GenKeyPair(keyType string) (PubKey, PrivKey, error) {
	switch keyType {
	case SecretFree:
		return nil, nil, nil
	case Sr25519:
		pub, priv := GenSrKey()
		return pub, priv, nil
	case Ed25519:
		pub, priv := GenEdKey()
		return pub, priv, nil
	case Secp256k1:
		pub, priv := GenSecpKey()
		return pub, priv, nil
	default:
		return nil, nil, NoKeyType
	}
}

// data: (keyTypeBytes + keyBytes)
func PubKeyFromBytes(data []byte) (PubKey, error) {
	if len(data) < KeyTypeBytLen {
		return nil, errors.New("null data")
	}
	keyTypeByt := data[:KeyTypeBytLen]
	switch string(keyTypeByt) {
	case SecretFreeIdx:
		return nil, nil
	case Sr25519Idx:
		return SrPubKeyFromBytes(data[KeyTypeBytLen:]), nil
	case Ed25519Idx:
		return EdPubKeyFromBytes(data[KeyTypeBytLen:]), nil
	case Secp256k1Idx:
		return SecpPubkeyFromBytes(data[KeyTypeBytLen:]), nil
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
