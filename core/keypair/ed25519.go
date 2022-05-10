package keypair

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	. "github.com/yu-org/yu/common"
)

// ----- Public Key ------

type EdPubkey struct {
	pubkey ed25519.PubKey
}

func EdPubKeyFromBytes(data []byte) *EdPubkey {
	return &EdPubkey{pubkey: data}
}

func (epb *EdPubkey) Address() Address {
	addressByt := epb.pubkey.Address().Bytes()
	return BytesToAddress(addressByt)
}

func (epb *EdPubkey) VerifySignature(msg, sig []byte) bool {
	return epb.pubkey.VerifySignature(msg, sig)
}

func (epb *EdPubkey) Type() string {
	return epb.pubkey.Type()
}

func (epb *EdPubkey) Equals(key Key) bool {
	edkey, ok := key.(*EdPubkey)
	if !ok {
		return false
	}
	return epb.pubkey.Equals(edkey.pubkey)
}

func (epb *EdPubkey) Bytes() []byte {
	return epb.pubkey.Bytes()
}

func (epb *EdPubkey) String() string {
	return ToHex(epb.Bytes())
}

func (epb *EdPubkey) BytesWithType() []byte {
	return append([]byte(Ed25519Idx), epb.pubkey.Bytes()...)
}

func (epb *EdPubkey) StringWithType() string {
	return ToHex(epb.BytesWithType())
}

// ------ Private Key -------

type EdPrivkey struct {
	privkey ed25519.PrivKey
}

func (epr *EdPrivkey) SignData(data []byte) ([]byte, error) {
	return epr.privkey.Sign(data)
}

func (epr *EdPrivkey) Type() string {
	return epr.privkey.Type()
}

func (epr *EdPrivkey) Equals(key Key) bool {
	edKey, ok := key.(*EdPrivkey)
	if !ok {
		return false
	}
	return epr.privkey.Equals(edKey.privkey)
}

func (epr *EdPrivkey) Bytes() []byte {
	return epr.privkey.Bytes()
}

func (epr *EdPrivkey) String() string {
	return ToHex(epr.Bytes())
}

func (epr *EdPrivkey) BytesWithType() []byte {
	return append([]byte(Ed25519Idx), epr.privkey.Bytes()...)
}

func (epr *EdPrivkey) StringWithType() string {
	return ToHex(epr.BytesWithType())
}

func GenEdKeyWithSecret(secret []byte) (PubKey, PrivKey) {
	edPrivKey := ed25519.GenPrivKeyFromSecret(secret)
	privkey := &EdPrivkey{edPrivKey}
	pubkey := &EdPubkey{edPrivKey.PubKey().(ed25519.PubKey)}
	return pubkey, privkey
}

func GenEdKey() (PubKey, PrivKey) {
	edPrivKey := ed25519.GenPrivKey()
	privkey := &EdPrivkey{edPrivKey}
	pubkey := &EdPubkey{edPrivKey.PubKey().(ed25519.PubKey)}
	return pubkey, privkey
}
