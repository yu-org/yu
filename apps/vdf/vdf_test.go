package vdf

import (
	"crypto/rand"
	"math/big"
	"testing"
)

func TestVDFBasic(t *testing.T) {
	N, err := GenerateRSAmodulus(1024)
	if err != nil {
		t.Fatal(err)
	}

	var x *big.Int
	for {
		r, _ := rand.Int(rand.Reader, N)
		if gcd(r, N).Cmp(big.NewInt(1)) == 0 {
			x = r
			break
		}
	}

	tVal := uint64(20)
	res, err := Eval(x, tVal, N, 128)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := Verify(x, res.Y, res.Pi, tVal, N, 128)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("proof verification failed")
	}
}
