package keypair

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	. "yu/common"
)

//type Ed25519 struct {
//	pubKey  ed25519.PubKey
//	privKey ed25519.PrivKey
//}

type EdPubkey struct {
	pubkey ed25519.PubKey
}

func (epb *EdPubkey) Address() Address {
	return epb.pubkey.Address().Bytes()
}

func (epb *EdPubkey) VerifySignature(msg, sig []byte) bool {
	return epb.pubkey.VerifySignature(msg, sig)
}

type EdPrivkey struct {
	privkey ed25519.PrivKey
}

func (epr *EdPrivkey) SignData(data []byte) ([]byte, error) {
	return epr.privkey.Sign(data)
}

func genEd25519() (*EdPubkey, *EdPrivkey) {
	edPrivKey := ed25519.GenPrivKey()
	privkey := &EdPrivkey{edPrivKey}
	pubkey := &EdPubkey{edPrivKey.PubKey().(ed25519.PubKey)}
	return pubkey, privkey
}
