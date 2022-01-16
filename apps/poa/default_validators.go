package poa

import (
	. "github.com/yu-org/yu/core/keypair"
)

type pair struct {
	pubkey  PubKey
	privkey PrivKey
}

func InitKeypair(idx int) (PubKey, PrivKey, []ValidatorAddrIp) {
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
	validatorsAddrs := []ValidatorAddrIp{
		{Addr: pub0.Address(), P2pIP: "12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
		{Addr: pub1.Address(), P2pIP: "12D3KooWSKPs95miv8wzj3fa5HkJ1tH7oEGumsEiD92n2MYwRtQG"},
		{Addr: pub2.Address(), P2pIP: "12D3KooWRuwP7nXaRhZrmoFJvPPGat2xPafVmGpQpZs5zKMtwqPH"},
	}

	return myPubkey, myPrivkey, validatorsAddrs
}
