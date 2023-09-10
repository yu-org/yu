package keypair

import (
	"github.com/tendermint/tendermint/crypto/secp256k1"
	. "github.com/yu-org/yu/common"
)

// ------ Public Key ------

type SecpPubkey struct {
	pubkey secp256k1.PubKey
}

func SecpPubkeyFromBytes(data []byte) *SecpPubkey {
	return &SecpPubkey{pubkey: data}
}

func (spb *SecpPubkey) Address() Address {
	addressByt := spb.pubkey.Address().Bytes()
	return BytesToAddress(addressByt)
}

func (spb *SecpPubkey) VerifySignature(msg, sig []byte) bool {
	return spb.pubkey.VerifySignature(msg, sig)
}

func (spb *SecpPubkey) Type() string {
	return spb.pubkey.Type()
}

func (spb *SecpPubkey) Equals(key Key) bool {
	srKey, ok := key.(*SecpPubkey)
	if !ok {
		return false
	}
	return spb.pubkey.Equals(srKey.pubkey)
}

func (spb *SecpPubkey) Bytes() []byte {
	return spb.pubkey.Bytes()
}

func (spb *SecpPubkey) String() string {
	return ToHex(spb.Bytes())
}

func (spb *SecpPubkey) BytesWithType() []byte {
	return append([]byte(Sr25519Idx), spb.pubkey.Bytes()...)
}

func (spb *SecpPubkey) StringWithType() string {
	return ToHex(spb.BytesWithType())
}

// ----- Private Key ------

type SecpPrivkey struct {
	privkey secp256k1.PrivKey
}

func SecpPrivkeyFromBytes(data []byte) *SecpPrivkey {
	return &SecpPrivkey{privkey: data}
}

func (spr *SecpPrivkey) SignData(data []byte) ([]byte, error) {
	return spr.privkey.Sign(data)
}

func (spr *SecpPrivkey) Type() string {
	return spr.privkey.Type()
}

func (spr *SecpPrivkey) Equals(key Key) bool {
	srKey, ok := key.(*SecpPrivkey)
	if !ok {
		return false
	}
	return spr.privkey.Equals(srKey.privkey)
}

func (spr *SecpPrivkey) Bytes() []byte {
	return spr.privkey.Bytes()
}

func (spr *SecpPrivkey) String() string {
	return ToHex(spr.Bytes())
}

func (spr *SecpPrivkey) BytesWithType() []byte {
	return append([]byte(Sr25519Idx), spr.privkey.Bytes()...)
}

func (spr *SecpPrivkey) StringWithType() string {
	return ToHex(spr.BytesWithType())
}

func (spr *SecpPrivkey) GenPubkey() []byte {
	return spr.privkey.PubKey().Bytes()
}

func GenSecpKeyWithSecret(secret []byte) (PubKey, PrivKey) {
	secpPrivkey := secp256k1.GenPrivKeySecp256k1(secret)
	privkey := &SecpPrivkey{secpPrivkey}
	pubkey := &SecpPubkey{secpPrivkey.PubKey().(secp256k1.PubKey)}
	return pubkey, privkey
}

func GenSecpKey() (PubKey, PrivKey) {
	secpPrivkey := secp256k1.GenPrivKey()
	privkey := &SecpPrivkey{secpPrivkey}
	pubkey := &SecpPubkey{secpPrivkey.PubKey().(secp256k1.PubKey)}
	return pubkey, privkey
}
