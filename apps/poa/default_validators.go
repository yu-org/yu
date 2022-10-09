package poa

import (
	. "github.com/yu-org/yu/core/keypair"
)

type pair struct {
	pubkey  PubKey
	privkey PrivKey
}

func InitDefaultKeypairs(idx int) (PubKey, PrivKey, []ValidatorInfo) {
	pub0, priv0 := GenSrKeyWithSecret([]byte("node1"))
	println("pubkey0: ", pub0.String())

	pub1, priv1 := GenSrKeyWithSecret([]byte("node2"))
	println("pubkey1: ", pub1.String())

	pub2, priv2 := GenSrKeyWithSecret([]byte("node3"))
	println("pubkey2: ", pub2.String())

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
	validatorsAddrs := []ValidatorInfo{
		{Pubkey: pub0, P2pID: "12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
		{Pubkey: pub1, P2pID: "12D3KooWSKPs95miv8wzj3fa5HkJ1tH7oEGumsEiD92n2MYwRtQG"},
		{Pubkey: pub2, P2pID: "12D3KooWRuwP7nXaRhZrmoFJvPPGat2xPafVmGpQpZs5zKMtwqPH"},
	}

	return myPubkey, myPrivkey, validatorsAddrs
}
