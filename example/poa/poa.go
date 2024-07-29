package main

import (
	"github.com/yu-org/yu/apps/asset"
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

	// myPubkey, myPrivkey, validatorsAddrs := poa.InitDefaultKeypairs(idx)
	poaConf := poa.DefaultCfg(idx)

	cfg := startup.InitKernelConfigFromPath("yu_conf/kernel.toml")
	startup.DefaultStartup(
		cfg,
		poa.NewPoa(poaConf),
		asset.NewAsset("YuCoin"),
	)
}
