package main

import (
	"github.com/yu-org/yu/apps/asset"
	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/keypair"
	"github.com/yu-org/yu/startup"
	"os"
	"strconv"
)

func main() {
	pubkey0, privkey0 := keypair.GenSrKey([]byte("yu"))
	pubkey1, privkey1 := keypair.GenSrKey([]byte("boyi"))
	pubkey2, privkey2 := keypair.GenSrKey([]byte("gaoyao"))

	pubkeyPool := []keypair.PubKey{
		pubkey0, pubkey1, pubkey2,
	}
	privkeyPool := []keypair.PrivKey{
		privkey0, privkey1, privkey2,
	}

	idx, err := strconv.Atoi(os.Args[0])
	if err != nil {
		panic(err)
	}

	// local-node keypair
	pubkey := pubkeyPool[idx]
	privkey := privkeyPool[idx]

	poaTripod := poa.NewPoa(1024, pubkey, privkey, pubkeyPool)
	startup.StartUp(poaTripod, asset.NewAsset("YuCoin"))
}
