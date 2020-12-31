package keypair

import (
	. "github.com/tendermint/tendermint/crypto/ed25519"
	. "yu/common"
)

type Ed25519 struct {
	pubKey  PubKey
	privKey PrivKey
}

func(ed *Ed25519) Address() Address {
	return ed.pubKey.Address().Bytes()
}

func (ed *Ed25519) SignData(data []byte) ([]byte, error) {
	return ed.privKey.Sign(data)
}

func (ed *Ed25519) VerifySigner(msg, sig []byte) bool {
	return ed.pubKey.VerifySignature(msg, sig)
}

func genEd25519() *Ed25519 {
	privKey := GenPrivKey()
	pubKey := privKey.PubKey()
	return &Ed25519{
		pubKey.(PubKey),
		privKey,
	}
}
