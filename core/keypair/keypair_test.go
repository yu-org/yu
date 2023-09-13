package keypair

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"testing"
)

func TestKey(t *testing.T) {
	testKey(t, Sr25519)
	testKey(t, Ed25519)
	testKey(t, Secp256k1)
}

func testKey(t *testing.T, keyType string) {
	t.Log("------- test ", keyType)
	pubkey, privkey, err := GenKeyPair(keyType)
	assert.NoError(t, err, "generate key failed")
	t.Logf("public key is %s", pubkey.String())
	t.Logf("private key is %s", privkey.String())
	wrCall := &WrCall{
		TripodName: "asset",
		FuncName:   "Transfer",
		Params:     "params-json-codec",
	}

	// just for print
	byt, err := json.Marshal(wrCall)
	assert.NoError(t, err)
	t.Logf("wrcall json: %s", byt)

	hash, err := wrCall.Hash()
	assert.NoError(t, err, "hash wrcall failed")
	signByt, err := privkey.SignData(hash)
	assert.NoError(t, err)
	signHex := ToHex(signByt)
	t.Logf("signature: %s, length: %d", signHex, len(signHex))

	ok := pubkey.VerifySignature(hash, signByt)
	assert.True(t, ok)
}
