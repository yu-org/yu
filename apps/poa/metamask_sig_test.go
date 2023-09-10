package poa

import (
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	"testing"
)

func TestSig(t *testing.T) {
	wrCall := common.WrCall{
		TripodName: "question",
		FuncName:   "AddQuestion",
		Params:     `{"title":"aaa","content":"bbb","timestamp":1692652052812}`,
		LeiPrice:   0,
		Tips:       0,
	}
	t.Log(wrCall)
	hash, err := wrCall.Hash()
	assert.NoError(t, err)
	t.Logf("hash is %x", hash)

	pub, priv := keypair.GenSecpKeyWithSecret([]byte("test"))
	t.Logf("raw privkey is %s, pubkey is %s", priv.String(), pub.String())
	sig, err := priv.SignData([]byte(MetamaskMsg(hash)))
	assert.NoError(t, err)
	t.Logf("sig is %x", sig)
}
