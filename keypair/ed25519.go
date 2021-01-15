package keypair

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	. "yu/common"
)

type EdPubkey struct {
	pubkey ed25519.PubKey
}

func (epb *EdPubkey) Address() Address {
	addressByt := epb.pubkey.Address().Bytes()
	return BytesToAddress(addressByt)
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
