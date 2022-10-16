package keypair

import (
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
	if err != nil {
		panic("generate key error: " + err.Error())
	}
	t.Logf("public key is %s", pubkey.String())
	t.Logf("private key is %s", privkey.String())
	ecall := &WrCall{
		TripodName: "asset",
		ExecName:   "Transfer",
		Params:     "params",
	}

	hash, err := ecall.Hash()
	if err != nil {
		t.Fatalf("hash error: %s", err.Error())
	}
	signByt, err := privkey.SignData(hash)
	if err != nil {
		panic("sign data error: " + err.Error())
	}
	ok := pubkey.VerifySignature(hash, signByt)
	assert.True(t, ok)
}
