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

// TestPubkeyTypeIndex pins each key type to the index it declares. The index is
// the only thing telling PubKeyFromBytes which curve a serialized key is on, so
// a key tagging itself as another type is not a cosmetic slip.
func TestPubkeyTypeIndex(t *testing.T) {
	for keyType, wantIdx := range map[string]string{
		Sr25519:   Sr25519Idx,
		Ed25519:   Ed25519Idx,
		Secp256k1: Secp256k1Idx,
	} {
		pubkey, privkey, err := GenKeyPair(keyType)
		assert.NoError(t, err, keyType)
		assert.Equal(t, wantIdx, string(pubkey.BytesWithType()[:KeyTypeBytLen]),
			"%s public key is tagged with the wrong type index", keyType)
		assert.Equal(t, wantIdx, string(privkey.BytesWithType()[:KeyTypeBytLen]),
			"%s private key is tagged with the wrong type index", keyType)
	}
}

// TestPubkeyRoundTrip checks that a public key serialized with its type comes
// back as the same key. PoA puts miner pubkeys in block headers this way and
// reads them back with PubKeyFromBytes, so a wrong type index decodes the key
// on the wrong curve: a different address, and signatures that never verify.
func TestPubkeyRoundTrip(t *testing.T) {
	for _, keyType := range []string{Sr25519, Ed25519, Secp256k1} {
		t.Run(keyType, func(t *testing.T) {
			pubkey, privkey, err := GenKeyPair(keyType)
			assert.NoError(t, err)

			decoded, err := PubKeyFromBytes(pubkey.BytesWithType())
			assert.NoError(t, err)
			assert.Equal(t, keyType, decoded.Type())
			assert.True(t, decoded.Equals(pubkey), "decoded key differs from the original")
			assert.Equal(t, pubkey.Address(), decoded.Address())

			fromStr, err := PubkeyFromStr(pubkey.StringWithType())
			assert.NoError(t, err)
			assert.True(t, fromStr.Equals(pubkey), "key decoded from hex differs from the original")

			// The decoded key is what verifies signatures in practice.
			hash, err := (&WrCall{TripodName: "asset", FuncName: "Transfer"}).Hash()
			assert.NoError(t, err)
			sig, err := privkey.SignData(hash)
			assert.NoError(t, err)
			assert.True(t, decoded.VerifySignature(hash, sig))
		})
	}
}
