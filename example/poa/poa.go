package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/asset"
	"github.com/yu-org/yu/apps/base"
	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/core/startup"
	"os"
	"strconv"
)

func main() {
	// dayu, _ := keypair.GenSrKey([]byte("yu"))
	// boyi, _ := keypair.GenSrKey([]byte("boyi"))

	idxStr := os.Args[1]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		panic(err)
	}

	myPubkey, myPrivkey, validatorsAddrs := poa.InitDefaultKeypairs(idx)

	logrus.Info("My Address is ", myPubkey.Address().String())
	startup.StartUp(
		base.NewBase(base.Full).Tripod,
		poa.NewPoa(myPubkey, myPrivkey, validatorsAddrs).Tripod,
		asset.NewAsset("YuCoin").Tripod,
	)
}
