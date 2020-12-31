package keypair

import (
	. "github.com/tendermint/tendermint/crypto/sr25519"
	. "yu/common"
)

type Sr25519 struct {
	pubKey  PubKey
	privKey PrivKey
}

func(sr *Sr25519) Address() Address {
	return sr.pubKey.Address().Bytes()
}

func (sr *Sr25519) SignData(data []byte) ([]byte, error) {
	return sr.privKey.Sign(data)
}

func (sr *Sr25519) VerifySigner(msg, sig []byte) bool {
	return sr.pubKey.VerifySignature(msg, sig)
}

func genSr25519() *Sr25519 {
	privKey := GenPrivKey()
	pubKey := privKey.PubKey()
	return &Sr25519{
		pubKey.(PubKey),
		privKey,
	}
}
