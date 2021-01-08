package keypair

import (
	"github.com/tendermint/tendermint/crypto/sr25519"
	. "yu/common"
)

type SrPubkey struct {
	pubkey sr25519.PubKey
}

func (spb *SrPubkey) Address() Address {
	return spb.pubkey.Address().Bytes()
}

func (spb *SrPubkey) VerifySignature(msg, sig []byte) bool {
	return spb.pubkey.VerifySignature(msg, sig)
}

type SrPrivkey struct {
	privkey sr25519.PrivKey
}

func (spr *SrPrivkey) SignData(data []byte) ([]byte, error) {
	return spr.privkey.Sign(data)
}

func genSr25519() (*SrPubkey, *SrPrivkey) {
	srPrivkey := sr25519.GenPrivKey()
	privkey := &SrPrivkey{srPrivkey}
	pubkey := &SrPubkey{srPrivkey.PubKey().(sr25519.PubKey)}
	return pubkey, privkey
}
