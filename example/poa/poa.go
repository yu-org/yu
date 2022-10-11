package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/asset"
	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/startup"
	"os"
	"strconv"
)

var secrets = []string{
	"node1",
	"node2",
	"node3",
}

func main() {
	// dayu, _ := keypair.GenSrKey([]byte("yu"))
	// boyi, _ := keypair.GenSrKey([]byte("boyi"))

	idxStr := os.Args[1]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		panic(err)
	}

	// myPubkey, myPrivkey, validatorsAddrs := poa.InitDefaultKeypairs(idx)
	poaConf := &poa.PoaConfig{
		KeyType:  keypair.Sr25519,
		MySecret: secrets[idx],
		Validators: []*poa.ValidatorConf{
			{Pubkey: "", P2pIp: "12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
			{Pubkey: "", P2pIp: "12D3KooWSKPs95miv8wzj3fa5HkJ1tH7oEGumsEiD92n2MYwRtQG"},
			{Pubkey: "", P2pIp: "12D3KooWRuwP7nXaRhZrmoFJvPPGat2xPafVmGpQpZs5zKMtwqPH"},
		},
	}

	var myPubkey keypair.PubKey
	for i, secret := range secrets {
		pub, _ := keypair.GenSrKeyWithSecret([]byte(secret))
		logrus.Infof("pub%d is %s", i, pub.String())
		poaConf.Validators[i].Pubkey = pub.StringWithType()
		if idx == i {
			myPubkey = pub
		}
	}

	logrus.Info("My Address is ", myPubkey.Address().String())

	startup.InitConfigFromPath("yu_conf/kernel.toml")
	startup.StartUpFullNode(
		poa.NewPoa(poaConf),
		asset.NewAsset("YuCoin"),
	)
}
