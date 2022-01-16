package poa

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"testing"
)

func TestCompeteLeader(t *testing.T) {
	myPubkey1, myPrivkey1, validators1 := InitKeypair(0)
	node1 := NewPoa(myPubkey1, myPrivkey1, validators1)
	t.Log("addr1 = ", myPubkey1.Address().String())

	myPubkey2, myPrivkey2, validators2 := InitKeypair(1)
	node2 := NewPoa(myPubkey2, myPrivkey2, validators2)
	t.Log("addr2 = ", myPubkey2.Address().String())

	myPubkey3, myPrivkey3, validators3 := InitKeypair(2)
	node3 := NewPoa(myPubkey3, myPrivkey3, validators3)
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
