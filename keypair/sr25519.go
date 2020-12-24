package keypair

import (
	. "github.com/tendermint/tendermint/crypto/sr25519"
)

type Sr25519 struct {
	pubKey  PubKey
	privKey PrivKey
}

func (sr *Sr25519) SignData(data []byte) ([]byte, error) {
	return sr.privKey.Sign(data)
}

func (sr *Sr25519) VerifySigner(msg, sig []byte) (bool, error) {
	return sr.pubKey.VerifySignature(msg, sig), nil
}

func generateSr25519() *Sr25519 {
	privKey := GenPrivKey()
	pubKey := privKey.PubKey()
	return &Sr25519{
		pubKey.(PubKey),
		privKey,
	}
}
