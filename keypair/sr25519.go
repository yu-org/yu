package keypair

import (
	"github.com/tendermint/tendermint/crypto/sr25519"
	. "yu/common"
)

// ------ Public Key ------

type SrPubkey struct {
	pubkey sr25519.PubKey
}

func srPubKeyFromBytes(data []byte) *SrPubkey {
	return &SrPubkey{pubkey: data}
}

func (spb *SrPubkey) Address() Address {
	addressByt := spb.pubkey.Address().Bytes()
	return BytesToAddress(addressByt)
}

func (spb *SrPubkey) VerifySignature(msg, sig []byte) bool {
	return spb.pubkey.VerifySignature(msg, sig)
}

func (spb *SrPubkey) Type() string {
	return spb.pubkey.Type()
}

func (spb *SrPubkey) Equals(key Key) bool {
	srKey, ok := key.(*SrPubkey)
	if !ok {
		return false
	}
	return spb.pubkey.Equals(srKey.pubkey)
}

func (spb *SrPubkey) Bytes() []byte {
	return spb.pubkey.Bytes()
}

func (spb *SrPubkey) String() string {
	return ToHex(spb.Bytes())
}

// ----- Private Key ------

type SrPrivkey struct {
	privkey sr25519.PrivKey
}

func (spr *SrPrivkey) SignData(data []byte) ([]byte, error) {
	return spr.privkey.Sign(data)
}

func (spr *SrPrivkey) Type() string {
	return spr.privkey.Type()
}

func (spr *SrPrivkey) Equals(key Key) bool {
	srKey, ok := key.(*SrPrivkey)
	if !ok {
		return false
	}
	return spr.privkey.Equals(srKey.privkey)
}

func (spr *SrPrivkey) Bytes() []byte {
	return spr.privkey.Bytes()
}

func (spr *SrPrivkey) String() string {
	return ToHex(spr.Bytes())
}

func genSr25519() (*SrPubkey, *SrPrivkey) {
	srPrivkey := sr25519.GenPrivKey()
	privkey := &SrPrivkey{srPrivkey}
	pubkey := &SrPubkey{srPrivkey.PubKey().(sr25519.PubKey)}
	return pubkey, privkey
}
