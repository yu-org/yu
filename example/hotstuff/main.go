package main

import (
	"github.com/yu-org/yu/apps/asset"
	"github.com/yu-org/yu/apps/hotstuff"
	"github.com/yu-org/yu/keypair"
	"github.com/yu-org/yu/startup"
	"os"
	"strconv"
)

type pair struct {
	pubkey  keypair.PubKey
	privkey keypair.PrivKey
}

func main() {
	pub0, priv0 := keypair.GenSrKey([]byte("node1"))
	pub1, priv1 := keypair.GenSrKey([]byte("node2"))
	pub2, priv2 := keypair.GenSrKey([]byte("node3"))

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

	idxStr := os.Args[0]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		panic(err)
	}
	myPubkey := pairArray[idx].pubkey
	myPrivkey := pairArray[idx].privkey

	validatorsMap := map[string]string{
		pub0.Address().String(): "12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu",
		pub1.Address().String(): "12D3KooWSKPs95miv8wzj3fa5HkJ1tH7oEGumsEiD92n2MYwRtQG",
		pub2.Address().String(): "12D3KooWRuwP7nXaRhZrmoFJvPPGat2xPafVmGpQpZs5zKMtwqPH",
	}

	startup.StartUp(hotstuff.NewHotstuff(myPubkey, myPrivkey, validatorsMap), asset.NewAsset("YuCoin"))
}
