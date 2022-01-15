package poa

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	"testing"
)

type pair struct {
	pubkey  PubKey
	privkey PrivKey
}

func initKeypair(idx int) (PubKey, PrivKey, map[Address]string) {
	pub0, priv0 := GenSrKey([]byte("node1"))
	pub1, priv1 := GenSrKey([]byte("node2"))
	pub2, priv2 := GenSrKey([]byte("node3"))

	pairArray := []pair{
		{
			pubkey:  pub0,
			privkey: priv0,
		},
		{
			pubkey:  pub1,
			privkey: priv1,
		},
		{
			pubkey:  pub2,
			privkey: priv2,
		},
	}

	myPubkey := pairArray[idx].pubkey
	myPrivkey := pairArray[idx].privkey
	validatorsMap := map[Address]string{
		pub0.Address(): "12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu",
		pub1.Address(): "12D3KooWSKPs95miv8wzj3fa5HkJ1tH7oEGumsEiD92n2MYwRtQG",
		pub2.Address(): "12D3KooWRuwP7nXaRhZrmoFJvPPGat2xPafVmGpQpZs5zKMtwqPH",
	}

	return myPubkey, myPrivkey, validatorsMap
}

func TestCompeteLeader(t *testing.T) {
	myPubkey1, myPrivkey1, validatorsMap1 := initKeypair(0)
	node1 := NewPoa(myPubkey1, myPrivkey1, validatorsMap1)
	t.Log("addr1 = ", myPubkey1.Address().String())

	myPubkey2, myPrivkey2, validatorsMap2 := initKeypair(1)
	node2 := NewPoa(myPubkey2, myPrivkey2, validatorsMap2)
	t.Log("addr2 = ", myPubkey2.Address().String())

	myPubkey3, myPrivkey3, validatorsMap3 := initKeypair(2)
	node3 := NewPoa(myPubkey3, myPrivkey3, validatorsMap3)
	t.Log("addr3 = ", myPubkey3.Address().String())

	for i := 1; i <= 30; i++ {
		bn := BlockNum(i)
		t.Log("block number = ", bn)
		addr1 := node1.CompeteLeader(bn)
		addr2 := node2.CompeteLeader(bn)
		addr3 := node3.CompeteLeader(bn)

		mod := (bn - 1) % 3
		switch mod {
		case 0:
			assert.Equal(t, myPubkey1.Address().String(), addr1.String(), "addr = %s", addr1.String())
		case 1:
			assert.Equal(t, myPubkey2.Address().String(), addr2.String(), "addr = %s", addr2.String())
		case 2:
			assert.Equal(t, myPubkey3.Address().String(), addr3.String(), "addr = %s", addr3.String())
		}
	}
}
