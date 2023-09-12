package poa

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	"testing"
)

var (
	wrCall = common.WrCall{
		TripodName: "question",
		FuncName:   "AddQuestion",
		Params:     `{"title":"aaa","content":"bbb","timestamp":1692652052812}`,
		LeiPrice:   0,
		Tips:       0,
	}
	pub, priv = keypair.GenSecpKeyWithSecret([]byte("test"))
)

func TestSig(t *testing.T) {
	t.Log(wrCall)
	hash, err := wrCall.Hash()
	assert.NoError(t, err)
	t.Logf("hash is %x", hash)

	t.Logf("raw privkey is %s, pubkey is %s", priv.String(), pub.String())
	metaMsg := MetamaskMsg(hash)
	t.Logf("%s", metaMsg)
	sig, err := priv.SignData(MetamaskMsg(hash))
	assert.NoError(t, err)
	t.Logf("sig is %s", common.ToHex(sig))
}

func TestVerify(t *testing.T) {

	prv, err := crypto.HexToECDSA("9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a09")
	assert.NoError(t, err)
	hash, err := wrCall.Hash()
	assert.NoError(t, err)

	metaMsg := MetamaskMsg(hash)
	t.Logf("  %v", metaMsg)

	mmMsgHash := crypto.Keccak256(metaMsg)
	t.Logf("metamask msg hash: %x", mmMsgHash)

	//// generate signature
	//sig, err := crypto.Sign(mmMsgHash, prv)
	//assert.NoError(t, err)
	//t.Logf("metamask sig: %x", sig)

	//sig := common.FromHex("0x3af49bd96fd526a6ce335eeac513d879dc68ab013b5fc4bde5297c8b9faebf822629d7f6c305125ed167bbd9a9f89a8ab56c28ced686007722a3f9ced4d250c41b")
	sig := common.FromHex("0x939082bc0ccf4a63dbbefafcf9c449af8de8595af5867db7251e992460f5cb54507c6e1f2f22f740404a595ca1227b5b0c9b22d9db094261e993ca5631fe6afb1b")
	t.Logf("signature bytes is %v", sig)
	pubkey, err := crypto.Ecrecover(mmMsgHash, sig)
	assert.NoError(t, err)
	assert.Equal(t, pubkey, crypto.FromECDSAPub(&prv.PublicKey))

	// assert.True(t, crypto.VerifySignature(pubkey, mmMsgHash, sig))
}
