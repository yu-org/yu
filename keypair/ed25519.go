package keypair

import (
	. "crypto/ed25519"
	"crypto/rand"
	"errors"
)

type Ed25519 struct {
	pubKey  PublicKey
	privKey PrivateKey
}

func (ed *Ed25519) SignData(data []byte) ([]byte, error) {
	if len(ed.privKey) != PrivateKeySize {
		return nil, errors.New("privkey size error")
	}
	return Sign(ed.privKey, data), nil
}

func (ed *Ed25519) VerifySigner(msg, sig []byte) (bool, error) {
	if len(ed.pubKey) != PublicKeySize {
		return false, errors.New("pubkey size error")
	}
	return Verify(ed.pubKey, msg, sig), nil
}

func generateEd25519() (*Ed25519, error) {
	pubKey, privKey, err := GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &Ed25519{
		pubKey,
		privKey,
	}, nil
}
